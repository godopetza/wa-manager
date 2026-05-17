package controllers

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/godopetza/wa-manager/initializers"
	"github.com/godopetza/wa-manager/models"
)

// GET /api/groups
func ListGroups(c *gin.Context) {
	var groups []models.ContactGroup
	initializers.DB.Preload("Contacts").Order("created_at desc").Find(&groups)
	c.JSON(http.StatusOK, groups)
}

// GET /api/groups/:id
func GetGroup(c *gin.Context) {
	var group models.ContactGroup
	if err := initializers.DB.Preload("Contacts").First(&group, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

type createGroupBody struct {
	Name string `json:"name" binding:"required"`
}

// POST /api/groups
func CreateGroup(c *gin.Context) {
	var body createGroupBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group := models.ContactGroup{Name: body.Name}
	if err := initializers.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

// DELETE /api/groups/:id
func DeleteGroup(c *gin.Context) {
	initializers.DB.Where("group_id = ?", c.Param("id")).Delete(&models.Contact{})
	initializers.DB.Delete(&models.ContactGroup{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}

type addContactBody struct {
	Phone string `json:"phone" binding:"required"`
	Name  string `json:"name"`
}

// POST /api/groups/:id/contacts
func AddContact(c *gin.Context) {
	var body addContactBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var groupID uint
	if err := initializers.DB.Model(&models.ContactGroup{}).Select("id").First(&groupID, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	contact := models.Contact{
		GroupID: groupID,
		Phone:   strings.TrimPrefix(body.Phone, "+"),
		Name:    body.Name,
	}
	if err := initializers.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

// DELETE /api/groups/:id/contacts/:contactId
func RemoveContact(c *gin.Context) {
	initializers.DB.Delete(&models.Contact{}, c.Param("contactId"))
	c.Status(http.StatusNoContent)
}

// POST /api/groups/:id/import  — CSV with columns: phone, name (header row required)
func ImportContacts(c *gin.Context) {
	var group models.ContactGroup
	if err := initializers.DB.First(&group, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CSV"})
		return
	}

	phoneIdx, nameIdx := -1, -1
	for i, h := range headers {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "phone":
			phoneIdx = i
		case "name":
			nameIdx = i
		}
	}
	if phoneIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have a 'phone' column"})
		return
	}

	var contacts []models.Contact
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) <= phoneIdx {
			continue
		}
		phone := strings.TrimPrefix(strings.TrimSpace(row[phoneIdx]), "+")
		if phone == "" {
			continue
		}
		name := ""
		if nameIdx >= 0 && len(row) > nameIdx {
			name = strings.TrimSpace(row[nameIdx])
		}
		contacts = append(contacts, models.Contact{
			GroupID: group.ID,
			Phone:   phone,
			Name:    name,
		})
	}

	if len(contacts) > 0 {
		initializers.DB.CreateInBatches(contacts, 100)
	}
	c.JSON(http.StatusOK, gin.H{"imported": len(contacts)})
}
