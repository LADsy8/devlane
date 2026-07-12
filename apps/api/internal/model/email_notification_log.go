package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailNotificationLog tracks notification emails queued/sent for audit.
// Maps to the existing email_notification_logs table.
type EmailNotificationLog struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReceiverID       uuid.UUID  `gorm:"type:uuid;not null" json:"receiver_id"`
	TriggeredByID    *uuid.UUID `gorm:"type:uuid" json:"triggered_by_id,omitempty"`
	Subject          string     `gorm:"type:text" json:"subject"`
	EntityIdentifier *uuid.UUID `gorm:"type:uuid" json:"entity_identifier,omitempty"`
	EntityName       string     `gorm:"type:varchar(255)" json:"entity_name,omitempty"`
	Data             JSONMap    `gorm:"type:jsonb;serializer:json" json:"data,omitempty"`
	ProcessedAt      *time.Time `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	SentAt           *time.Time `gorm:"type:timestamptz" json:"sent_at,omitempty"`    // When queued to RabbitMQ
	Entity           string     `gorm:"type:varchar(200)" json:"entity,omitempty"`    // Legacy
	OldValue         string     `gorm:"type:varchar(300)" json:"old_value,omitempty"` // Legacy
	NewValue         string     `gorm:"type:varchar(300)" json:"new_value,omitempty"` // Legacy
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (EmailNotificationLog) TableName() string { return "email_notification_logs" }

func (e *EmailNotificationLog) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
