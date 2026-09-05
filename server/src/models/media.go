package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Media struct {
	Common
	Archivable
	Name       string         `gorm:"column:name;type:text;not null" json:"name"`
	FileName   string         `gorm:"column:file_name;type:text;not null" json:"fileName"`
	URL        string         `gorm:"column:url;type:text;not null" json:"url"`
	MimeType   string         `gorm:"column:mime_type;type:text;not null" json:"mimeType"`
	Size       int64          `gorm:"column:size;not null" json:"size"`
	Alt        string         `gorm:"column:alt;type:text" json:"alt"`
	Caption    string         `gorm:"column:caption;type:text" json:"caption"`
	UploadedBy string         `gorm:"column:uploaded_by;type:text;not null" json:"uploadedBy"`
	WorkspaceID *string       `gorm:"column:workspace_id;type:text;index" json:"workspaceId,omitempty"`
	Metadata   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Media) TableName() string { return "media" }
