package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Webhook struct {
	Common
	Archivable
	WorkspaceID     string         `gorm:"column:workspace_id;type:text;index;not null" json:"workspaceId"`
	Provider        string         `gorm:"column:provider;type:text;not null" json:"provider"`
	URL             string         `gorm:"column:url;type:text;not null" json:"url"`
	Secret          string         `gorm:"column:secret;type:text" json:"-"`
	Events          datatypes.JSON `gorm:"column:events;type:jsonb" json:"events"`
	Active          bool           `gorm:"column:active;not null;default:true" json:"active"`
	LastTriggeredAt *time.Time     `gorm:"column:last_triggered_at" json:"lastTriggeredAt,omitempty"`
	FailureCount    int            `gorm:"column:failure_count;not null;default:0" json:"failureCount"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Webhook) TableName() string { return "webhooks" }

type WebhookDelivery struct {
	Common
	WebhookID   string `gorm:"column:webhook_id;type:text;index;not null" json:"webhookId"`
	Event       string `gorm:"column:event;type:text;not null" json:"event"`
	Status      int    `gorm:"column:status;not null" json:"status"`
	RequestBody string `gorm:"column:request_body;type:text" json:"-"`
	ResponseBody string `gorm:"column:response_body;type:text" json:"-"`
	Duration    int    `gorm:"column:duration;not null;default:0" json:"duration"`
	Error       string `gorm:"column:error;type:text" json:"error,omitempty"`
}

func (WebhookDelivery) TableName() string { return "webhook_deliveries" }
