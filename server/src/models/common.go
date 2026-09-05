package models

import "time"

type Common struct {
	ID        string    `gorm:"column:id;type:text;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

type Archivable struct {
	ArchivedAt *time.Time `gorm:"column:archived_at" json:"archivedAt,omitempty"`
}
