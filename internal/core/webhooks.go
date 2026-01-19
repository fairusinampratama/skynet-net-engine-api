package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"skynet-net-engine-api/pkg/logger"
	"time"
	"go.uber.org/zap"
)

// WebhookPayload sent to the Brain
type WebhookPayload struct {
	Event    string      `json:"event"`     // e.g. "router.up", "router.down"
	RouterID int         `json:"router_id"`
	Host     string      `json:"host"`
	Data     interface{} `json:"data,omitempty"`
	Timestamp string     `json:"timestamp"`
}

// SendWebhook dispatches events to a configured URL
func SendWebhook(event string, routerID int, host string, data interface{}) {
	// Load from env or default
	targetURL := "http://localhost:8000/api/webhooks/net-engine" 
	if url := os.Getenv("WEBHOOK_URL"); url != "" {
		targetURL = url
	}

	payload := WebhookPayload{
		Event:     event,
		RouterID:  routerID,
		Host:      host,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	go func() {
		jsonData, _ := json.Marshal(payload)
		resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Warn("Failed to send webhook", zap.String("event", event), zap.Error(err))
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 400 {
			logger.Warn("Webhook server returned error", zap.Int("status", resp.StatusCode))
		} else {
			logger.Info("Webhook sent successfully", zap.String("event", event))
		}
	}()
}
