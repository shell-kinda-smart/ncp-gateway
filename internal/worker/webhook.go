package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type WebhookPayload struct {
	Event     string  `json:"event"`      // e.g. "deposit_detected"
	TxID      string  `json:"tx_id"`
	Coin      string  `json:"coin"`
	Address   string  `json:"address"`
	UserID    string  `json:"user_id,omitempty"`    // "from whom received" context
	Amount    float64 `json:"amount"`               // Numeric for machine parsing
	AmountStr string  `json:"amount_formatted"`     // Human readable "0.0019 LTC"
	USDAmount float64 `json:"usd_amount,omitempty"` // Real-time USD value
	USDPrice  float64 `json:"usd_price,omitempty"`  // Price at time of tx
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

func SendWebhook(url string, payload WebhookPayload) {
	if url == "" {
		return
	}

	go func() {
		// Serialize Payload to JSON first
		jsonBytes, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal webhook payload: %v", err)
			return
		}

		var body []byte
		if strings.Contains(url, "discord.com/api/webhooks") {
			// User requested "JSON not embed". 
			// We wraps the JSON string in a Discord 'content' field with code blocks.
			discordBody := map[string]string{
				"content": fmt.Sprintf("```json\n%s\n```", string(jsonBytes)),
			}
			body, _ = json.Marshal(discordBody)
		} else {
			// Standard JSON for everything else
			body = jsonBytes
		}

		if err != nil {
			log.Printf("Failed to marshal webhook body: %v", err)
			return
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Failed to send webhook to %s: %v", url, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Printf("Webhook receiver returned error: %d", resp.StatusCode)
		} else {
			log.Printf("Webhook sent successfully to %s", url)
		}
	}()
}
