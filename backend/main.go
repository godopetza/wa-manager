package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/godopetza/wa-manager/controllers"
	"github.com/godopetza/wa-manager/initializers"
	"github.com/godopetza/wa-manager/middleware"
	"github.com/godopetza/wa-manager/models"
)

func init() {
	initializers.ConnectToDB()
	initializers.DB.AutoMigrate(
		&models.WATemplate{},
		&models.ContactGroup{},
		&models.Contact{},
		&models.Campaign{},
		&models.CampaignMessage{},
	)
}

func main() {
	r := gin.Default()

	cfg := cors.DefaultConfig()
	cfg.AllowOriginFunc = func(origin string) bool {
		allowed := []string{
			os.Getenv("FRONTEND_URL"),
			"http://localhost:3000",
		}
		for _, a := range allowed {
			if a != "" && origin == a {
				return true
			}
		}
		return false
	}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	cfg.AllowCredentials = false
	cfg.MaxAge = 12 * time.Hour
	r.Use(cors.New(cfg))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	api := r.Group("/api")

	// ── Auth (public) ─────────────────────────────────────────────────────────
	api.POST("/auth/login", controllers.Login)

	// ── Meta webhooks (public — verified by HMAC signature) ───────────────────
	webhooks := api.Group("/webhooks")
	{
		webhooks.GET("/whatsapp", controllers.WAWebhookVerify)
		webhooks.POST("/whatsapp", controllers.WAWebhookReceive)
	}

	// ── Protected routes ──────────────────────────────────────────────────────
	auth := api.Group("/", middleware.RequireAuth())
	{
		// Templates
		auth.GET("/templates", controllers.ListTemplates)
		auth.POST("/templates", controllers.CreateTemplate)
		auth.GET("/templates/:id", controllers.GetTemplate)
		auth.DELETE("/templates/:id", controllers.DeleteTemplate)
		auth.POST("/templates/:id/submit", controllers.SubmitTemplate)
		auth.POST("/templates/:id/sync", controllers.SyncTemplate)
		auth.POST("/templates/:id/upload-image", controllers.UploadTemplateImage)

		// Contact groups
		auth.GET("/groups", controllers.ListGroups)
		auth.POST("/groups", controllers.CreateGroup)
		auth.GET("/groups/:id", controllers.GetGroup)
		auth.DELETE("/groups/:id", controllers.DeleteGroup)
		auth.POST("/groups/:id/contacts", controllers.AddContact)
		auth.DELETE("/groups/:id/contacts/:contactId", controllers.RemoveContact)
		auth.POST("/groups/:id/import", controllers.ImportContacts)

		// Campaigns
		auth.GET("/campaigns", controllers.ListCampaigns)
		auth.POST("/campaigns", controllers.CreateCampaign)
		auth.GET("/campaigns/:id", controllers.GetCampaign)
		auth.DELETE("/campaigns/:id", controllers.DeleteCampaign)
		auth.POST("/campaigns/:id/send", controllers.SendCampaign)
		auth.GET("/campaigns/:id/messages", controllers.ListCampaignMessages)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("wa-manager backend listening on :%s", port)
	r.Run(":" + port)
}
