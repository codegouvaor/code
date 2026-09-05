package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Article struct {
	Common
	Archivable
	Title            string         `gorm:"column:title;type:text;not null" json:"title"`
	Slug             string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	Excerpt          string         `gorm:"column:excerpt;type:text" json:"excerpt"`
	Content          string         `gorm:"column:content;type:text" json:"content"`
	Type             string         `gorm:"column:type;type:text;not null;default:'article'" json:"type"`
	Status           string         `gorm:"column:status;type:text;not null;default:'draft'" json:"status"`
	CategoryID       *string        `gorm:"column:category_id;type:text;index" json:"categoryId,omitempty"`
	Team             string         `gorm:"column:team;type:text" json:"team"`
	AuthorID         string         `gorm:"column:author_id;type:text;index;not null" json:"authorId"`
	PublishedAt      *time.Time     `gorm:"column:published_at" json:"publishedAt,omitempty"`
	ScheduledAt      *time.Time     `gorm:"column:scheduled_at" json:"scheduledAt,omitempty"`
	Tags             datatypes.JSON `gorm:"column:tags;type:jsonb" json:"tags"`
	SeoTitle         string         `gorm:"column:seo_title;type:text" json:"seoTitle"`
	SeoDescription   string         `gorm:"column:seo_description;type:text" json:"seoDescription"`
	SeoOgImage       string         `gorm:"column:seo_og_image;type:text" json:"seoOgImage"`
	Priority         string         `gorm:"column:priority;type:text;default:'normal'" json:"priority"`
	Channel          string         `gorm:"column:channel;type:text" json:"channel"`
	Views            int64          `gorm:"column:views;not null;default:0" json:"views"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Author   *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

func (Article) TableName() string { return "articles" }

type ArticleTag struct {
	ArticleID string `gorm:"column:article_id;type:text;primaryKey" json:"articleId"`
	TagID     string `gorm:"column:tag_id;type:text;primaryKey" json:"tagId"`
}

func (ArticleTag) TableName() string { return "article_tags" }
