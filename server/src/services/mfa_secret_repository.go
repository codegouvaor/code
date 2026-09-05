package services

import (
	"context"

	"github.com/codegouvaor/code/server/src/models"
	"gorm.io/gorm"
)

type mfaSecretRepository struct{ db *gorm.DB }

func (r *mfaSecretRepository) GetByUserID(ctx context.Context, userID string) (*models.MfaSecret, error) {
	var item models.MfaSecret
	err := r.db.WithContext(ctx).First(&item, "user_id = ?", userID).Error
	return &item, err
}

func (r *mfaSecretRepository) Create(ctx context.Context, secret *models.MfaSecret) error {
	return r.db.WithContext(ctx).Create(secret).Error
}

func (r *mfaSecretRepository) Update(ctx context.Context, secret *models.MfaSecret) error {
	return r.db.WithContext(ctx).Save(secret).Error
}

func (r *mfaSecretRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.MfaSecret{}).Error
}
