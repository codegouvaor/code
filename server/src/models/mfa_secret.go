package models

import "time"

type MfaSecret struct {
	ID          string     `gorm:"type:text;primaryKey" json:"id"`
	UserID      string     `gorm:"column:user_id;type:text;uniqueIndex;not null" json:"userId"`
	Secret      string     `gorm:"column:secret;type:text;not null" json:"-"`
	OtpauthUrl  *string    `gorm:"column:otpauth_url;type:text" json:"otpauthUrl,omitempty"`
	QrCodeUrl   *string    `gorm:"column:qr_code_url;type:text" json:"qrCodeUrl,omitempty"`
	ConfirmedAt *time.Time `gorm:"column:confirmed_at" json:"confirmedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updatedAt"`
}

func (MfaSecret) TableName() string {
	return "mfa_secrets"
}
