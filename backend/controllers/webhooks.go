package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/godopetza/wa-manager/initializers"
	"github.com/godopetza/wa-manager/models"
)

// GET /api/webhooks/whatsapp — Meta hub.verify_token handshake.
func WAWebhookVerify(c *gin.Context) {
	if c.Query("hub.mode") != "subscribe" {
		c.Status(http.StatusForbidden)
		return
	}
	token := c.Query("hub.verify_token")
	if token == "" {
		c.Status(http.StatusForbidden)
		return
	}
	if shared := os.Getenv("META_WEBHOOK_VERIFY_TOKEN"); shared != "" && token == shared {
		c.String(http.StatusOK, c.Query("hub.challenge"))
		return
	}
	c.Status(http.StatusForbidden)
}

// POST /api/webhooks/whatsapp — receive delivery status updates from Meta.
// Updates CampaignMessage rows so the history page shows real delivery state.
func WAWebhookReceive(c *gin.Context) {
	rawBody, _ := io.ReadAll(c.Request.Body)
	sig := c.GetHeader("X-Hub-Signature-256")
	if !verifyMetaSignature(rawBody, sig) {
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	processDeliveryStatuses(payload)
	c.Status(http.StatusOK)
}

// verifyMetaSignature checks X-Hub-Signature-256 using HMAC-SHA256.
// Tries META_APP_SECRET then OLD_META_APP_SECRET for zero-downtime key rotation.
func verifyMetaSignature(body []byte, sig string) bool {
	if len(sig) < 7 || sig[:7] != "sha256=" {
		return false
	}
	for _, envKey := range []string{"META_APP_SECRET", "OLD_META_APP_SECRET"} {
		secret := os.Getenv(envKey)
		if secret == "" {
			continue
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return true
		}
	}
	return false
}

// processDeliveryStatuses walks the Meta webhook payload and updates
// CampaignMessage rows by wamid (WAMessageID).
func processDeliveryStatuses(payload map[string]interface{}) {
	entries, _ := payload["entry"].([]interface{})
	for _, e := range entries {
		entry, _ := e.(map[string]interface{})
		changes, _ := entry["changes"].([]interface{})
		for _, ch := range changes {
			change, _ := ch.(map[string]interface{})
			value, _ := change["value"].(map[string]interface{})
			statuses, _ := value["statuses"].([]interface{})
			for _, s := range statuses {
				status, _ := s.(map[string]interface{})
				wamid, _ := status["id"].(string)
				statusStr, _ := status["status"].(string)
				if wamid == "" || statusStr == "" {
					continue
				}

				var msg models.CampaignMessage
				if err := initializers.DB.Where("wa_message_id = ?", wamid).First(&msg).Error; err != nil {
					continue // not our message
				}

				now := time.Now()
				updates := map[string]interface{}{"status": statusStr}
				switch statusStr {
				case "sent":
					updates["sent_at"] = now
				case "delivered":
					updates["delivered_at"] = now
				case "read":
					updates["read_at"] = now
				case "failed":
					if errors, ok := status["errors"].([]interface{}); ok && len(errors) > 0 {
						if errObj, ok := errors[0].(map[string]interface{}); ok {
							if title, ok := errObj["title"].(string); ok {
								updates["fail_reason"] = title
							}
						}
					}
				}

				if err := initializers.DB.Model(&msg).Updates(updates).Error; err != nil {
					log.Printf("[webhook] update CampaignMessage %d: %v", msg.ID, err)
				}
			}
		}
	}
}
