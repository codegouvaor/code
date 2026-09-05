package models

import "time"

type Simulcast struct {
	Common
	AnimeID       string     `gorm:"column:anime_id;type:text;index;not null" json:"animeId"`
	AirDay        string     `gorm:"column:air_day;type:text;not null" json:"airDay"`
	AirTime       string     `gorm:"column:air_time;type:text;not null" json:"airTime"`
	AirTimezone   string     `gorm:"column:air_timezone;type:text;not null" json:"airTimezone"`
	Region        string     `gorm:"column:region;type:text;not null;default:'global'" json:"region"`
	Platform      string     `gorm:"column:platform;type:text" json:"platform"`
	EpisodeNumber int        `gorm:"column:episode_number;default:0" json:"episodeNumber"`
	Season        int        `gorm:"column:season;default:0" json:"season"`
	StartsAt      time.Time  `gorm:"column:starts_at;not null" json:"startsAt"`
	EndsAt        *time.Time `gorm:"column:ends_at" json:"endsAt,omitempty"`
	Status        string     `gorm:"column:status;type:text;not null;default:'scheduled'" json:"status"`
}

func (Simulcast) TableName() string { return "simulcasts" }

type ReleaseSchedule struct {
	Common
	AnimeID     string    `gorm:"column:anime_id;type:text;index;not null" json:"animeId"`
	EpisodeID   *string   `gorm:"column:episode_id;type:text" json:"episodeId,omitempty"`
	ScheduledAt time.Time `gorm:"column:scheduled_at;not null" json:"scheduledAt"`
	Region      string    `gorm:"column:region;type:text;not null;default:'global'" json:"region"`
	Status      string    `gorm:"column:status;type:text;not null;default:'scheduled'" json:"status"`
}

func (ReleaseSchedule) TableName() string { return "release_schedules" }
