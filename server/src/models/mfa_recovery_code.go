package models

import "time"

type MfaRecoveryCode struct {
	ID        string     `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	UserID    string     `gorm:"column:user_id;type:text;not null;index" json:"userId"`
	CodeHash  string     `gorm:"column:code_hash;type:text;not null" json:"-"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"-"`
}

func (MfaRecoveryCode) TableName() string {
	return "mfa_recovery_codes"
}
