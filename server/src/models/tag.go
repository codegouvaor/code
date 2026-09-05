package models

type Tag struct {
	Common
	Name string `gorm:"column:name;type:text;not null" json:"name"`
	Slug string `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
}

func (Tag) TableName() string { return "tags" }
