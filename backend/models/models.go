package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WATemplate struct {
	gorm.Model
	Name         string         `gorm:"uniqueIndex;not null"`
	Language     string         `gorm:"default:'en_US'"`
	Category     string         `gorm:"not null"` // MARKETING, UTILITY, AUTHENTICATION
	Components   datatypes.JSON // Meta components array JSON
	MetaStatus   string         `gorm:"default:'draft'"` // draft, pending, approved, rejected
	MetaID       string
	RejectReason string
	Campaigns    []Campaign `gorm:"foreignKey:TemplateID"`
}

type ContactGroup struct {
	gorm.Model
	Name      string     `gorm:"uniqueIndex;not null"`
	Contacts  []Contact  `gorm:"foreignKey:GroupID"`
	Campaigns []Campaign `gorm:"foreignKey:GroupID"`
}

type Contact struct {
	gorm.Model
	GroupID uint   `gorm:"not null;index"`
	Phone   string `gorm:"not null"` // E.164 without +
	Name    string
	Tags    datatypes.JSON // {"key":"value"} for template variable substitution
}

type Campaign struct {
	gorm.Model
	Name        string
	TemplateID  uint
	Template    WATemplate
	GroupID     uint
	Group       ContactGroup
	ImageURL    string // optional hosted image URL for template header
	Status      string `gorm:"default:'draft'"` // draft, sending, completed, partial_failed
	SentCount   int
	FailCount   int
	StartedAt   *time.Time
	CompletedAt *time.Time
	Messages    []CampaignMessage `gorm:"foreignKey:CampaignID"`
}

// CampaignMessage tracks one WhatsApp message per contact per campaign.
// WAMessageID is the wamid returned by Meta — used to correlate delivery webhooks.
type CampaignMessage struct {
	gorm.Model
	CampaignID  uint `gorm:"not null;index"`
	ContactID   uint `gorm:"not null"`
	Contact     Contact
	Phone       string
	Status      string `gorm:"default:'pending'"` // pending, sent, delivered, read, failed
	WAMessageID string `gorm:"index"`             // wamid from Meta, used for webhook correlation
	FailReason  string
	SentAt      *time.Time
	DeliveredAt *time.Time
	ReadAt      *time.Time
}
