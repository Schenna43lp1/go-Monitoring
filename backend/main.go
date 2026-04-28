package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Payload struct {
	AgentID   string         `json:"agent_id"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type Store struct {
	mu    sync.Mutex
	items []Payload
}

func main() {
	port := env("BACKEND_PORT", "8080")
	store := &Store{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", withCORS(healthHandler))
	mux.HandleFunc("/api/data", withCORS(dataHandler(store)))

	addr := ":" + port
	log.Printf("backend listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func dataHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			store.mu.Lock()
			items := append([]Payload(nil), store.items...)
			store.mu.Unlock()
			writeJSON(w, http.StatusOK, items)

		case http.MethodPost:
			var payload Payload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if payload.Timestamp == "" {
				payload.Timestamp = time.Now().UTC().Format(time.RFC3339)
			}

			store.mu.Lock()
			store.items = append(store.items, payload)
			store.mu.Unlock()

			writeJSON(w, http.StatusCreated, payload)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
