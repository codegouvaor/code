package models

import (
	"time"

	"gorm.io/datatypes"
)

type Notification struct {
	Common
	UserID   string         `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Type     string         `gorm:"column:type;type:text;not null" json:"type"`
	Title    string         `gorm:"column:title;type:text;not null" json:"title"`
	Body     string         `gorm:"column:body;type:text" json:"body"`
	Link     string         `gorm:"column:link;type:text" json:"link"`
	Read     bool           `gorm:"column:read;default:false" json:"read"`
	ReadAt   *time.Time     `gorm:"column:read_at" json:"readAt,omitempty"`
	Metadata datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
}

func (Notification) TableName() string { return "notifications" }

type NotificationTemplate struct {
	Common
	Type     string         `gorm:"column:type;type:text;uniqueIndex;not null" json:"type"`
	Subject  string         `gorm:"column:subject;type:text;not null" json:"subject"`
	Body     string         `gorm:"column:body;type:text;not null" json:"body"`
	IsHtml   bool           `gorm:"column:is_html;default:false" json:"isHtml"`
	Metadata datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
}

func (NotificationTemplate) TableName() string { return "notification_templates" }
