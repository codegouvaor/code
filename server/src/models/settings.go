package models

import "gorm.io/datatypes"

type SystemSetting struct {
	Common
	Key         string         `gorm:"column:key;type:text;uniqueIndex;not null" json:"key"`
	Value       datatypes.JSON `gorm:"column:value;type:jsonb" json:"value"`
	Category    string         `gorm:"column:category;type:text;not null" json:"category"`
	Description string         `gorm:"column:description;type:text" json:"description"`
}

func (SystemSetting) TableName() string { return "system_settings" }

type SeoMeta struct {
	Common
	PagePath     string `gorm:"column:page_path;type:text;uniqueIndex;not null" json:"pagePath"`
	Title        string `gorm:"column:title;type:text" json:"title"`
	Description  string `gorm:"column:description;type:text" json:"description"`
	Keywords     string `gorm:"column:keywords;type:text" json:"keywords"`
	OgImage      string `gorm:"column:og_image;type:text" json:"ogImage"`
	CanonicalUrl string `gorm:"column:canonical_url;type:text" json:"canonicalUrl"`
}

func (SeoMeta) TableName() string { return "seo_meta" }
