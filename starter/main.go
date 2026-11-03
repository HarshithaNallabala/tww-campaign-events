
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	events   = make(map[string]Event)
	eventsMu sync.Mutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /events", handlePostEvents)
	mux.HandleFunc("GET /campaigns/{campaignID}/stats", handleGetStats)

	addr := ":8080"
	fmt.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// handlePostEvents ingests a JSON array of events.
func handlePostEvents(w http.ResponseWriter, r *http.Request) {
	var incoming []json.RawMessage

	err := json.NewDecoder(r.Body).Decode(&incoming)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	accepted := 0
	duplicates := 0
	rejected := 0

	eventsMu.Lock()
	defer eventsMu.Unlock()

	for _, rawEvent := range incoming {
		var event Event

		if err := json.Unmarshal(rawEvent, &event); err != nil {
			rejected++
			continue
		}

		if !isValidEvent(event) {
			rejected++
			continue
		}

		if _, exists := events[event.EventID]; exists {
			duplicates++
			continue
		}

		events[event.EventID] = event
		accepted++
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"accepted":   accepted,
		"duplicates": duplicates,
		"rejected":   rejected,
	})
}

// handleGetStats returns aggregated stats for one campaign.
func handleGetStats(w http.ResponseWriter, r *http.Request) {
	campaignID := r.PathValue("campaignID")

	stats := map[string]int{
		"sent":      0,
		"delivered": 0,
		"opened":    0,
		"clicked":   0,
	}

	eventsMu.Lock()

	for _, event := range events {
		if event.CampaignID != campaignID {
			continue
		}

		switch event.Type {
		case "sent":
			stats["sent"]++
		case "delivered":
			stats["delivered"]++
		case "opened":
			stats["opened"]++
		case "clicked":
			stats["clicked"]++
		}
	}

	eventsMu.Unlock()

	writeJSON(w, http.StatusOK, stats)
}

func isValidEvent(event Event) bool {
    if strings.TrimSpace(event.EventID) == "" ||
        strings.TrimSpace(event.CampaignID) == "" ||
        strings.TrimSpace(event.ContactID) == "" {
        return false
    }

    if !isValidType(event.Type) {
        return false
    }

    // Validate timestamp format.
    _, err := time.Parse(time.RFC3339, event.Timestamp)
    return err == nil
}

func isValidType(eventType string) bool {
	switch eventType {
	case "sent", "delivered", "opened", "clicked":
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
