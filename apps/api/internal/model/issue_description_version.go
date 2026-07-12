package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IssueDescriptionVersion is an immutable snapshot of a work item's description,
// recorded whenever the description changes so users can browse history and
// restore an earlier version. Matches the issue_description_versions table.
type IssueDescriptionVersion struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IssueID         uuid.UUID  `gorm:"type:uuid;not null" json:"issue_id"`
	DescriptionHTML string     `gorm:"column:description_html;type:text" json:"description_html"`
	CreatedByID     *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	OwnedByID       *uuid.UUID `gorm:"type:uuid" json:"owned_by_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (IssueDescriptionVersion) TableName() string { return "issue_description_versions" }

func (v *IssueDescriptionVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
