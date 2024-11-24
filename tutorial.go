package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Description string `json:"shortDescription"`
	Price       string `json:"price"`
}

type Receipt struct {
	ID          string `json:"id,omitempty"`
	Retailer    string `json:"retailer"`
	Date        string `json:"purchaseDate"`
	Time        string `json:"purchaseTime"`
	Items       []Item `json:"items"`
	TotalAmount string `json:"total"`
}

type Points struct {
	Points int `json:"points"`
}

var store = make(map[string]Receipt)

func main() {
	http.HandleFunc("/receipts/process", handleReceipt)
	// Keep this route as the base
	http.HandleFunc("/receipts/", calculatePointsWithPoints) // New route to handle /points at the end
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var rcp Receipt
	if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	rcp.ID = uuid.New().String()
	store[rcp.ID] = rcp
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": rcp.ID})
}

func calculatePointsWithPoints(w http.ResponseWriter, r *http.Request) {
	// Extract the receipt ID and ensure the correct path format
	id := strings.TrimPrefix(r.URL.Path, "/receipts/")
	id = strings.TrimSuffix(id, "/points") // Remove the "/points" part

	rcp, found := store[id]
	if !found {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	points := computePoints(rcp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Points{Points: points})
}

func computePoints(r Receipt) int {
	points := 0

	points += countAlphanumeric(r.Retailer)
	if total, _ := parsePrice(r.TotalAmount); total == math.Floor(total) {
		points += 50
	}
	if total, _ := parsePrice(r.TotalAmount); math.Mod(total, 0.25) == 0 {
		points += 25
	}
	points += 5 * (len(r.Items) / 2)

	for _, item := range r.Items {
		if len(strings.TrimSpace(item.Description))%3 == 0 {
			price, _ := parsePrice(item.Price)
			points += int(math.Ceil(price * 0.2))
		}
	}

	if date, _ := time.Parse("2006-01-02", r.Date); date.Day()%2 != 0 {
		points += 6
	}

	if t, _ := time.Parse("15:04", r.Time); t.Hour() >= 14 && t.Hour() < 16 {
		points += 10
	}

	return points
}

func countAlphanumeric(s string) int {
	count := 0
	for _, ch := range s {
		if isAlphaNumeric(ch) {
			count++
		}
	}
	return count
}

func isAlphaNumeric(ch rune) bool {
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ('0' <= ch && ch <= '9')
}

func parsePrice(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
