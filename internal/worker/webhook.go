package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type WebhookPayload struct {
	TxID      string `json:"tx_id"`
	Coin      string `json:"coin"`
	Amount    string `json:"amount"`
	ToAddress string `json:"to_address"`
	Status    string `json:"status"`
}

func SendWebhook(url string, payload WebhookPayload) {
	if url == "" {
		return
	}

	go func() {
		// Use a retry mechanism or queue in production. For now, direct send.
		var body []byte
		var err error

		if strings.Contains(url, "discord.com/api/webhooks") {
			// Format for Discord
			discordBody := map[string]interface{}{
				"username": "Crypto Gateway",
				"embeds": []map[string]interface{}{
					{
						"title":       "💰 New Deposit Detected!",
						"color":       5763719, // Green
						"description": fmt.Sprintf("A new **%s** deposit has been detected.", payload.Coin),
						"fields": []map[string]interface{}{
							{"name": "Amount", "value": payload.Amount, "inline": true},
							{"name": "Coin", "value": payload.Coin, "inline": true},
							{"name": "Address", "value": payload.ToAddress},
							{"name": "Transaction ID", "value": payload.TxID},
							{"name": "Status", "value": payload.Status},
						},
						"timestamp": time.Now().Format(time.RFC3339),
					},
				},
			}
			body, err = json.Marshal(discordBody)
		} else {
			// Standard JSON
			body, err = json.Marshal(payload)
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
