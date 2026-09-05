package models

import "gorm.io/gorm"

type Studio struct {
	Common
	Name        string         `gorm:"column:name;type:text;uniqueIndex;not null" json:"name"`
	Slug        string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	LogoUrl     string         `gorm:"column:logo_url;type:text" json:"logoUrl"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Website     string         `gorm:"column:website;type:text" json:"website"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Studio) TableName() string { return "studios" }
