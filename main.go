package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type AgentData struct {
	AgentID   string         `json:"agent_id"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Meta      map[string]any `json:"meta,omitempty"`
}

func main() {
	backendURL := env("BACKEND_URL", "http://localhost:8080/ingest")
	port := env("AGENT_PORT", "9090")
	client := &http.Client{Timeout: 10 * time.Second}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data AgentData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if data.Timestamp == "" {
			data.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		status, err := sendToBackend(r.Context(), client, backendURL, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"backend_status": status,
		})
	})

	http.HandleFunc("/sample", func(w http.ResponseWriter, r *http.Request) {
		data := AgentData{
			AgentID:   env("AGENT_ID", "agent-1"),
			Message:   "hello from go agent",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Meta: map[string]any{
				"hostname": hostname(),
			},
		}

		status, err := sendToBackend(r.Context(), client, backendURL, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"backend_status": status,
			"sent":           data,
		})
	})

	addr := ":" + port
	log.Printf("agent listening on %s, backend=%s", addr, backendURL)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func sendToBackend(ctx context.Context, client *http.Client, backendURL string, data AgentData) (int, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendURL, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send to backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}
