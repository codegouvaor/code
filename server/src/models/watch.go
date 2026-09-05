package models

import "time"

type WatchProgress struct {
	Common
	UserID      string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	EpisodeID   string    `gorm:"column:episode_id;type:text;index;not null" json:"episodeId"`
	AnimeID     string    `gorm:"column:anime_id;type:text;index;not null" json:"animeId"`
	Progress    float64   `gorm:"column:progress;type:real;default:0" json:"progress"`
	Duration    float64   `gorm:"column:duration;type:real;default:0" json:"duration"`
	Percentage  float64   `gorm:"column:percentage;type:real;default:0" json:"percentage"`
	Completed   bool      `gorm:"column:completed;default:false" json:"completed"`
	LastWatched time.Time `gorm:"column:last_watched;not null" json:"lastWatched"`
}

func (WatchProgress) TableName() string { return "watch_progress" }

type WatchHistory struct {
	Common
	UserID    string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	EpisodeID string    `gorm:"column:episode_id;type:text;index;not null" json:"episodeId"`
	AnimeID   string    `gorm:"column:anime_id;type:text;index;not null" json:"animeId"`
	WatchedAt time.Time `gorm:"column:watched_at;not null" json:"watchedAt"`
	Duration  float64   `gorm:"column:duration;type:real" json:"duration"`
}

func (WatchHistory) TableName() string { return "watch_history" }
