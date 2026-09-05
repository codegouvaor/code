package models

import "time"

type Schedule struct {
	Common
	EntityType   string     `gorm:"column:entity_type;type:text;not null" json:"entityType"`
	EntityID     string     `gorm:"column:entity_id;type:text;not null" json:"entityId"`
	Title        string     `gorm:"column:title;type:text;not null" json:"title"`
	ScheduledAt  time.Time  `gorm:"column:scheduled_at;not null" json:"scheduledAt"`
	Status       string     `gorm:"column:status;type:text;not null;default:'pending'" json:"status"`
	PublishedAt  *time.Time `gorm:"column:published_at" json:"publishedAt,omitempty"`
	CancelledAt  *time.Time `gorm:"column:cancelled_at" json:"cancelledAt,omitempty"`
	WorkspaceID  *string    `gorm:"column:workspace_id;type:text;index" json:"workspaceId,omitempty"`
	CreatedBy    string     `gorm:"column:created_by;type:text;not null" json:"createdBy"`
}

func (Schedule) TableName() string { return "schedules" }
