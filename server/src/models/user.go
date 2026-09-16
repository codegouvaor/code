package models

import (
	"time"
)

type User struct {
	Common
	// Username is the Code handle used by the platform routes
	// (/[owner]/[repo]). It is nullable so that accounts created before the
	// platform layer still work; the API refuses to expose a profile without
	// one and lets the user claim it explicitly.
	Username          *string    `gorm:"column:username;type:text;uniqueIndex" json:"username,omitempty"`
	Email             string     `gorm:"column:email;type:text;not null" json:"email"`
	EmailNormalized   string     `gorm:"column:email_normalized;type:text;uniqueIndex" json:"-"`
	DisplayName       string     `gorm:"column:display_name;type:text;not null" json:"displayName"`
	AvatarURL         *string    `gorm:"column:avatar_url;type:text" json:"avatarUrl,omitempty"`
	Status            string     `gorm:"column:status;type:text;not null;default:'active'" json:"status"`
	EmailVerifiedAt   *time.Time `gorm:"column:email_verified_at" json:"emailVerifiedAt,omitempty"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at" json:"passwordChangedAt,omitempty"`
	DisabledAt        *time.Time `gorm:"column:disabled_at" json:"disabledAt,omitempty"`
}

func (User) TableName() string {
	return "users"
}
