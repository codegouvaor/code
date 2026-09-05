package services

import (
	"context"

	"time"

	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
	"gorm.io/gorm"
)

type mfaRecoveryCodeRepository struct{ db *gorm.DB }

func (r *mfaRecoveryCodeRepository) Create(ctx context.Context, code *models.MfaRecoveryCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *mfaRecoveryCodeRepository) CreateBatch(ctx context.Context, codes []*models.MfaRecoveryCode) error {
	if len(codes) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&codes).Error
}

func (r *mfaRecoveryCodeRepository) GetByUserID(ctx context.Context, userID string) ([]models.MfaRecoveryCode, error) {
	var items []models.MfaRecoveryCode
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *mfaRecoveryCodeRepository) GetByID(ctx context.Context, id string) (*models.MfaRecoveryCode, error) {
	var item models.MfaRecoveryCode
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "MFA_RECOVERY_CODE_NOT_FOUND", "The requested MFA recovery code was not found.", nil))
}

func (r *mfaRecoveryCodeRepository) MarkUsed(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.MfaRecoveryCode{}).Where("id = ?", id).Update("used_at", now).Error
}

func (r *mfaRecoveryCodeRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.MfaRecoveryCode{}).Error
}
