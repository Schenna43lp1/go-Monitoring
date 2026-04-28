package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Payload struct {
	AgentID   string         `json:"agent_id"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type Config struct {
	AgentID    string
	BackendURL string
	Interval   time.Duration
}

func main() {
	cfg := Config{
		AgentID:    env("AGENT_ID", defaultAgentID()),
		BackendURL: env("BACKEND_URL", "http://localhost:8080/api/data"),
		Interval:   time.Duration(envInt("AGENT_INTERVAL_SECONDS", 30)) * time.Second,
	}

	client := &http.Client{Timeout: 10 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("agent=%s backend=%s interval=%s", cfg.AgentID, cfg.BackendURL, cfg.Interval)

	sendHeartbeat(ctx, client, cfg, "agent started")

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sendHeartbeat(context.Background(), client, cfg, "agent stopped")
			log.Println("agent stopped")
			return
		case <-ticker.C:
			sendHeartbeat(ctx, client, cfg, "agent heartbeat")
		}
	}
}

func sendHeartbeat(ctx context.Context, client *http.Client, cfg Config, message string) {
	payload := Payload{
		AgentID:   cfg.AgentID,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Meta: map[string]any{
			"hostname": hostname(),
			"ips":      localIPs(),
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
		},
	}

	status, err := postJSON(ctx, client, cfg.BackendURL, payload)
	if err != nil {
		log.Printf("send failed: %v", err)
		return
	}

	log.Printf("sent heartbeat status=%d", status)
}

func postJSON(ctx context.Context, client *http.Client, url string, payload Payload) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("post backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, fmt.Errorf("backend returned %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
}

func defaultAgentID() string {
	name := hostname()
	if name == "unknown" {
		return "agent-unknown"
	}
	return "agent-" + strings.ToLower(name)
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown"
	}
	return name
}

func localIPs() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	ips := make([]string, 0)
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		ips = append(ips, ip.String())
	}
	return ips
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
