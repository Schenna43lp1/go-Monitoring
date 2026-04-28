package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type AlertConfig struct {
	Enabled           bool    `json:"enabled"`
	DiscordWebhookURL string  `json:"discord_webhook_url"`
	CPUPercent        float64 `json:"cpu_percent"`
	RAMPercent        float64 `json:"ram_percent"`
	DiskPercent       float64 `json:"disk_percent"`
	CooldownSeconds   int     `json:"cooldown_seconds"`
}

type Store struct {
	mu            sync.Mutex
	items         []Payload
	alertConfig   AlertConfig
	lastAlertSent map[string]time.Time
	webhookClient *http.Client
}

func main() {
	port := env("BACKEND_PORT", "8080")
	store := &Store{
		alertConfig: AlertConfig{
			CPUPercent:      90,
			RAMPercent:      90,
			DiskPercent:     90,
			CooldownSeconds: 300,
		},
		lastAlertSent: make(map[string]time.Time),
		webhookClient: &http.Client{Timeout: 10 * time.Second},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", withCORS(healthHandler))
	mux.HandleFunc("/api/data", withCORS(dataHandler(store)))
	mux.HandleFunc("/api/alerts", withCORS(alertsHandler(store)))

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
			alertConfig := store.alertConfig
			store.mu.Unlock()

			store.checkAlerts(payload, alertConfig)

			writeJSON(w, http.StatusCreated, payload)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func alertsHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			store.mu.Lock()
			config := store.alertConfig
			store.mu.Unlock()
			writeJSON(w, http.StatusOK, config)

		case http.MethodPost:
			var config AlertConfig
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if config.CooldownSeconds <= 0 {
				config.CooldownSeconds = 300
			}

			store.mu.Lock()
			store.alertConfig = config
			store.mu.Unlock()

			writeJSON(w, http.StatusOK, config)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (store *Store) checkAlerts(payload Payload, config AlertConfig) {
	if !config.Enabled || config.DiscordWebhookURL == "" || payload.Meta == nil {
		return
	}

	checks := []struct {
		key       string
		label     string
		threshold float64
	}{
		{key: "cpu_percent", label: "CPU", threshold: config.CPUPercent},
		{key: "ram_percent", label: "RAM", threshold: config.RAMPercent},
		{key: "disk_percent", label: "Disk", threshold: config.DiskPercent},
	}

	for _, check := range checks {
		value, ok := numberFromMeta(payload.Meta, check.key)
		if !ok || check.threshold <= 0 || value < check.threshold {
			continue
		}

		alertKey := payload.AgentID + ":" + check.key
		if store.isCoolingDown(alertKey, time.Duration(config.CooldownSeconds)*time.Second) {
			continue
		}

		message := fmt.Sprintf(
			"Alert: %s auf %s ist bei %.1f%% und damit ueber %.1f%%.",
			check.label,
			payload.AgentID,
			value,
			check.threshold,
		)
		if err := sendDiscordWebhook(store.webhookClient, config.DiscordWebhookURL, message); err != nil {
			log.Printf("discord webhook failed: %v", err)
			continue
		}
		store.markAlertSent(alertKey)
	}
}

func (store *Store) isCoolingDown(key string, cooldown time.Duration) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	last, ok := store.lastAlertSent[key]
	return ok && time.Since(last) < cooldown
}

func (store *Store) markAlertSent(key string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.lastAlertSent[key] = time.Now()
}

func sendDiscordWebhook(client *http.Client, webhookURL string, content string) error {
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}
	return nil
}

func numberFromMeta(meta map[string]any, key string) (float64, bool) {
	value, ok := meta[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
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
