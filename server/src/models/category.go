package models

import "gorm.io/gorm"

type Category struct {
	Common
	Archivable
	Name        string         `gorm:"column:name;type:text;not null" json:"name"`
	Slug        string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Color       string         `gorm:"column;color;type:text" json:"color"`
	SortOrder   int            `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	ParentID    *string        `gorm:"column:parent_id;type:text;index" json:"parentId,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Category) TableName() string { return "categories" }
