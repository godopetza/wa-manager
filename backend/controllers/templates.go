package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/godopetza/wa-manager/initializers"
	"github.com/godopetza/wa-manager/models"
	"github.com/godopetza/wa-manager/services"
	"gorm.io/datatypes"
)

// GET /api/templates
func ListTemplates(c *gin.Context) {
	var templates []models.WATemplate
	initializers.DB.Order("created_at desc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

// GET /api/templates/:id
func GetTemplate(c *gin.Context) {
	var t models.WATemplate
	if err := initializers.DB.First(&t, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

type createTemplateBody struct {
	Name       string                   `json:"name" binding:"required"`
	Language   string                   `json:"language" binding:"required"`
	Category   string                   `json:"category" binding:"required"`
	Components []map[string]interface{} `json:"components" binding:"required"`
	SubmitNow  bool                     `json:"submitNow"` // true = submit to Meta immediately
}

// POST /api/templates
// Creates the template locally and optionally submits to Meta for approval.
func CreateTemplate(c *gin.Context) {
	var body createTemplateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	componentsJSON, _ := json.Marshal(body.Components)
	t := models.WATemplate{
		Name:       body.Name,
		Language:   body.Language,
		Category:   body.Category,
		Components: datatypes.JSON(componentsJSON),
		MetaStatus: "draft",
	}

	if body.SubmitNow {
		wabaID := os.Getenv("META_WABA_ID")
		metaID, err := services.SubmitTemplateToMeta(wabaID, body.Name, body.Language, body.Category, body.Components)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "meta submission failed: " + err.Error()})
			return
		}
		t.MetaID = metaID
		t.MetaStatus = "pending"
	}

	if err := initializers.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// POST /api/templates/:id/submit — submit a draft template to Meta.
func SubmitTemplate(c *gin.Context) {
	var t models.WATemplate
	if err := initializers.DB.First(&t, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var components []map[string]interface{}
	json.Unmarshal([]byte(t.Components), &components)

	wabaID := os.Getenv("META_WABA_ID")
	metaID, err := services.SubmitTemplateToMeta(wabaID, t.Name, t.Language, t.Category, components)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	initializers.DB.Model(&t).Updates(map[string]interface{}{
		"meta_id":     metaID,
		"meta_status": "pending",
	})
	c.JSON(http.StatusOK, t)
}

// POST /api/templates/:id/sync — pull latest approval status from Meta.
func SyncTemplate(c *gin.Context) {
	var t models.WATemplate
	if err := initializers.DB.First(&t, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	wabaID := os.Getenv("META_WABA_ID")
	all, err := services.SyncTemplatesFromMeta(wabaID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	for _, m := range all {
		if m["name"] == t.Name {
			status, _ := m["status"].(string)
			reason, _ := m["rejected_reason"].(string)
			updates := map[string]interface{}{"meta_status": status}
			if reason != "" {
				updates["reject_reason"] = reason
			}
			initializers.DB.Model(&t).Updates(updates)
			initializers.DB.First(&t, t.ID)
			break
		}
	}
	c.JSON(http.StatusOK, t)
}

// POST /api/templates/:id/upload-image
// Uploads the header image to Meta and returns the handle for template creation.
type uploadImageBody struct {
	ImageURL string `json:"imageUrl" binding:"required"`
}

func UploadTemplateImage(c *gin.Context) {
	var body uploadImageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handle, err := services.UploadImageToMeta(body.ImageURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"handle": handle})
}

// DELETE /api/templates/:id
func DeleteTemplate(c *gin.Context) {
	if err := initializers.DB.Delete(&models.WATemplate{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
