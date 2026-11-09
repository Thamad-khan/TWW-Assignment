package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Stats struct {
	CampaignID  string `json:"campaign_id"`
	Sent        int    `json:"sent"`
	Delivered   int    `json:"delivered"`
	Opened      int    `json:"opened"`
	Clicked     int    `json:"clicked"`
	UniqueOpens int    `json:"unique_opens"`
}

type Store struct {
	mu     sync.RWMutex
	events map[string]Event
	opened map[string]bool
	stats  map[string]*Stats
}

var db = Store{
	events: make(map[string]Event),
	opened: make(map[string]bool),
	stats:  make(map[string]*Stats),
}

type BatchResult struct {
	Accepted   int      `json:"accepted"`
	Duplicates int      `json:"duplicates"`
	Rejected   int      `json:"rejected"`
	Errors     []string `json:"errors"`
}

func valid(e Event) error {
	if strings.TrimSpace(e.EventID) == "" ||
		strings.TrimSpace(e.CampaignID) == "" ||
		strings.TrimSpace(e.ContactID) == "" {
		return &validationError{"event_id, campaign_id and contact_id are required"}
	}
	switch e.Type {
	case "sent", "delivered", "opened", "clicked":
	default:
		return &validationError{"invalid event type"}
	}
	if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		return &validationError{"timestamp must be RFC3339"}
	}
	return nil
}

type validationError struct{ message string }

func (e *validationError) Error() string { return e.message }

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid or oversized request body", http.StatusBadRequest)
		return
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "expected a JSON array of events", http.StatusBadRequest)
		return
	}

	result := BatchResult{Errors: []string{}}

	for i, item := range raw {
		var e Event
		if err := json.Unmarshal(item, &e); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, "item "+itoa(i)+": invalid JSON event")
			continue
		}
		if err := valid(e); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, "item "+itoa(i)+": "+err.Error())
			continue
		}

		db.mu.Lock()
		if _, exists := db.events[e.EventID]; exists {
			result.Duplicates++
			db.mu.Unlock()
			continue
		}

		db.events[e.EventID] = e
		s := db.stats[e.CampaignID]
		if s == nil {
			s = &Stats{CampaignID: e.CampaignID}
			db.stats[e.CampaignID] = s
		}

		switch e.Type {
		case "sent":
			s.Sent++
		case "delivered":
			s.Delivered++
		case "opened":
			s.Opened++
			key := e.CampaignID + "\x00" + e.ContactID
			if !db.opened[key] {
				db.opened[key] = true
				s.UniqueOpens++
			}
		case "clicked":
			s.Clicked++
		}
		db.mu.Unlock()
		result.Accepted++
	}

	w.Header().Set("Content-Type", "application/json")
	if result.Rejected > 0 {
		w.WriteHeader(http.StatusMultiStatus)
	}
	json.NewEncoder(w).Encode(result)
}

func itoa(i int) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.Trim(strings.ReplaceAll(
		string(rune(i+48)), "\x00", ""), " "), " ", ""))
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/campaigns/")
	if !strings.HasSuffix(path, "/stats") {
		http.NotFound(w, r)
		return
	}
	id := strings.TrimSuffix(path, "/stats")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	db.mu.RLock()
	s := db.stats[id]
	var result Stats
	if s != nil {
		result = *s
	} else {
		result = Stats{CampaignID: id}
	}
	db.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"status":"ok"}`)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", eventsHandler)
	mux.HandleFunc("/campaigns/", statsHandler)
	mux.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Campaign Events API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// Keep bytes imported for compatibility with future request validation.
var _ = bytes.MinRead
var _ = sort.Strings
