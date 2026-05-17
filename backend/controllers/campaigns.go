package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/godopetza/wa-manager/initializers"
	"github.com/godopetza/wa-manager/models"
	"github.com/godopetza/wa-manager/services"
)

// GET /api/campaigns
func ListCampaigns(c *gin.Context) {
	var campaigns []models.Campaign
	initializers.DB.Preload("Template").Preload("Group").Order("created_at desc").Find(&campaigns)
	c.JSON(http.StatusOK, campaigns)
}

// GET /api/campaigns/:id
func GetCampaign(c *gin.Context) {
	var campaign models.Campaign
	if err := initializers.DB.Preload("Template").Preload("Group").Preload("Messages.Contact").First(&campaign, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

type createCampaignBody struct {
	Name       string `json:"name" binding:"required"`
	TemplateID uint   `json:"templateId" binding:"required"`
	GroupID    uint   `json:"groupId" binding:"required"`
	ImageURL   string `json:"imageUrl"` // optional rendered card URL (go-invite-render integration)
}

// POST /api/campaigns
func CreateCampaign(c *gin.Context) {
	var body createCampaignBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign := models.Campaign{
		Name:       body.Name,
		TemplateID: body.TemplateID,
		GroupID:    body.GroupID,
		ImageURL:   body.ImageURL,
		Status:     "draft",
	}
	if err := initializers.DB.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, campaign)
}

// POST /api/campaigns/:id/send — fire the campaign to all contacts in the group.
// Sends concurrently (up to 5 goroutines) and records per-message results.
func SendCampaign(c *gin.Context) {
	var campaign models.Campaign
	if err := initializers.DB.Preload("Template").Preload("Group.Contacts").First(&campaign, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if campaign.Status == "sending" {
		c.JSON(http.StatusConflict, gin.H{"error": "campaign is already sending"})
		return
	}

	now := time.Now()
	initializers.DB.Model(&campaign).Updates(map[string]interface{}{
		"status":     "sending",
		"started_at": now,
	})

	contacts := campaign.Group.Contacts
	if len(contacts) == 0 {
		initializers.DB.Model(&campaign).Update("status", "completed")
		c.JSON(http.StatusOK, gin.H{"message": "no contacts in group"})
		return
	}

	// Build per-message rows so we can track delivery via webhook.
	messages := make([]models.CampaignMessage, len(contacts))
	for i, ct := range contacts {
		messages[i] = models.CampaignMessage{
			CampaignID: campaign.ID,
			ContactID:  ct.ID,
			Phone:      ct.Phone,
			Status:     "pending",
		}
	}
	initializers.DB.CreateInBatches(messages, 100)

	// Resolve component variables — if contact has tags, interpolate them.
	var rawComponents []map[string]interface{}
	json.Unmarshal([]byte(campaign.Template.Components), &rawComponents)

	go func() {
		sem := make(chan struct{}, 5) // max 5 concurrent sends
		var wg sync.WaitGroup
		sent, failed := 0, 0
		var mu sync.Mutex

		for i, ct := range contacts {
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int, contact models.Contact) {
				defer wg.Done()
				defer func() { <-sem }()

				msgID := messages[idx].ID
				phone := "+" + contact.Phone

				// If campaign has an image URL, send image first (go-invite-render card).
				if campaign.ImageURL != "" {
					services.SendImageWithToken(
						os.Getenv("META_PHONE_NUMBER_ID"),
						os.Getenv("META_ACCESS_TOKEN"),
						phone,
						campaign.ImageURL,
						"",
					)
				}

				componentsJSON, _ := json.Marshal(rawComponents)
				wamid, err := services.SendTemplateMessage(
					phone,
					campaign.Template.Name,
					campaign.Template.Language,
					string(componentsJSON),
				)

				now := time.Now()
				mu.Lock()
				if err != nil {
					failed++
					initializers.DB.Model(&models.CampaignMessage{}).Where("id = ?", msgID).Updates(map[string]interface{}{
						"status":      "failed",
						"fail_reason": err.Error(),
					})
					log.Printf("[campaign %d] send to %s failed: %v", campaign.ID, phone, err)
				} else {
					sent++
					initializers.DB.Model(&models.CampaignMessage{}).Where("id = ?", msgID).Updates(map[string]interface{}{
						"status":        "sent",
						"wa_message_id": wamid,
						"sent_at":       now,
					})
				}
				mu.Unlock()
			}(i, ct)
		}

		wg.Wait()
		completedAt := time.Now()
		status := "completed"
		if failed > 0 && sent == 0 {
			status = "failed"
		} else if failed > 0 {
			status = "partial_failed"
		}
		initializers.DB.Model(&campaign).Updates(map[string]interface{}{
			"status":       status,
			"sent_count":   sent,
			"fail_count":   failed,
			"completed_at": completedAt,
		})
		log.Printf("[campaign %d] done — sent=%d failed=%d", campaign.ID, sent, failed)
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "campaign started", "contacts": len(contacts)})
}

// GET /api/campaigns/:id/messages — paginated per-message delivery log.
func ListCampaignMessages(c *gin.Context) {
	var messages []models.CampaignMessage
	initializers.DB.Preload("Contact").
		Where("campaign_id = ?", c.Param("id")).
		Order("created_at asc").
		Find(&messages)
	c.JSON(http.StatusOK, messages)
}

// DELETE /api/campaigns/:id
func DeleteCampaign(c *gin.Context) {
	initializers.DB.Where("campaign_id = ?", c.Param("id")).Delete(&models.CampaignMessage{})
	initializers.DB.Delete(&models.Campaign{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}
