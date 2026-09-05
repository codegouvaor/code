package models

type SeoConfig struct {
	Common
	PagePath     string `gorm:"column:page_path;type:text;uniqueIndex;not null" json:"pagePath"`
	Title        string `gorm:"column:title;type:text;not null" json:"title"`
	Description  string `gorm:"column:description;type:text" json:"description"`
	OgImage      string `gorm:"column:og_image;type:text" json:"ogImage"`
	Canonical    string `gorm:"column:canonical;type:text" json:"canonical"`
	NoIndex      bool   `gorm:"column:no_index;not null;default:false" json:"noIndex"`
	Keywords     string `gorm:"column:keywords;type:text" json:"keywords"`
	Locale       string `gorm:"column:locale;type:text;default:'fr-FR'" json:"locale"`
}

func (SeoConfig) TableName() string { return "seo_configs" }
