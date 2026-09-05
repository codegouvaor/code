package models

import "time"

type NewsletterSubscriber struct {
	Common
	Email            string     `gorm:"column:email;type:text;not null" json:"email"`
	EmailNormalized  string     `gorm:"column:email_normalized;type:text;uniqueIndex;not null" json:"-"`
	Status           string     `gorm:"column:status;type:text;not null;default:'active'" json:"status"`
	WorkspaceID      *string    `gorm:"column:workspace_id;type:text;index" json:"workspaceId,omitempty"`
	SubscribedAt     time.Time  `gorm:"column:subscribed_at;not null" json:"subscribedAt"`
	UnsubscribedAt   *time.Time `gorm:"column:unsubscribed_at" json:"unsubscribedAt,omitempty"`
	UnsubscribeToken string     `gorm:"column:unsubscribe_token;type:text;uniqueIndex" json:"-"`
}

func (NewsletterSubscriber) TableName() string { return "newsletter_subscribers" }
