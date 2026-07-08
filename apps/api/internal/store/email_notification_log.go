package store

import (
	"context"
	"time"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailNotificationLogStore struct {
	db *gorm.DB
}

func NewEmailNotificationLogStore(db *gorm.DB) *EmailNotificationLogStore {
	return &EmailNotificationLogStore{db: db}
}

// Create inserts a single email notification log entry.
func (s *EmailNotificationLogStore) Create(ctx context.Context, log *model.EmailNotificationLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

// MarkSent updates sent_at to indicate the email was queued to RabbitMQ.
func (s *EmailNotificationLogStore) MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	return s.db.WithContext(ctx).
		Model(&model.EmailNotificationLog{}).
		Where("id = ?", id).
		Update("sent_at", sentAt).Error
}
