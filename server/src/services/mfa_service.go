package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/codegouvaor/info/server/src/config"
	"github.com/codegouvaor/info/server/src/interfaces"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
)

type MfaService struct {
	cfg   config.AuthConfig
	db    interfaces.Database
	repos *Repositories
}

type MfaSetupResult struct {
	Secret        string   `json:"secret"`
	OtpauthURL    string   `json:"otpauthUrl"`
	RecoveryCodes []string `json:"recoveryCodes"`
}

type MfaVerifyResult struct {
	Verified bool `json:"verified"`
}

type MfaDisableResult struct {
	Disabled bool `json:"disabled"`
}

type MfaValidateLoginResult struct {
	Valid bool `json:"valid"`
}

type MfaRecoveryCodesResult struct {
	RecoveryCodes []string `json:"recoveryCodes"`
}

func NewMfaService(cfg config.AuthConfig, db interfaces.Database, repos *Repositories) *MfaService {
	return &MfaService{
		cfg:   cfg,
		db:    db,
		repos: repos,
	}
}

func (s *MfaService) Setup(ctx context.Context, userID string) (*MfaSetupResult, error) {
	user, err := s.repos.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err == nil && existing.ConfirmedAt != nil {
		return nil, utils.NewError(400, "MFA_ALREADY_ENABLED", "MFA is already enabled on this account.", nil)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.TOTPIssuer,
		AccountName: user.Email,
	})
	if err != nil {
		return nil, utils.NewError(500, "MFA_SETUP_FAILED", "Failed to generate MFA secret.", nil)
	}

	secret := key.Secret()
	now := time.Now().UTC()
	otpauthURL := key.URL()

	if existing != nil {
		existing.Secret = secret
		existing.OtpauthUrl = &otpauthURL
		existing.ConfirmedAt = nil
		existing.UpdatedAt = now
		if err := s.repos.MfaSecrets().Update(ctx, existing); err != nil {
			return nil, err
		}
	} else {
		mfaSecret := &models.MfaSecret{
			ID:         utils.NewID(),
			UserID:     userID,
			Secret:     secret,
			OtpauthUrl: &otpauthURL,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.repos.MfaSecrets().Create(ctx, mfaSecret); err != nil {
			return nil, err
		}
	}

	recoveryCodes := s.generateRecoveryCodes(s.cfg.MFARecoveryCodeLength)
	if err := s.storeRecoveryCodes(ctx, userID, recoveryCodes); err != nil {
		return nil, err
	}

	return &MfaSetupResult{
		Secret:        secret,
		OtpauthURL:    key.URL(),
		RecoveryCodes: recoveryCodes,
	}, nil
}

func (s *MfaService) VerifyAndEnable(ctx context.Context, userID string, code string) (*MfaVerifyResult, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return nil, utils.NewError(400, "MFA_NOT_SETUP", "MFA has not been set up. Call /auth/mfa/setup first.", nil)
	}
	if mfaSecret.ConfirmedAt != nil {
		return nil, utils.NewError(400, "MFA_ALREADY_ENABLED", "MFA is already enabled on this account.", nil)
	}

	valid := totp.Validate(code, mfaSecret.Secret)
	if !valid {
		return nil, utils.NewError(400, "INVALID_MFA_CODE", "The provided MFA code is invalid.", nil)
	}

	now := time.Now().UTC()
	mfaSecret.ConfirmedAt = &now
	mfaSecret.UpdatedAt = now
	if err := s.repos.MfaSecrets().Update(ctx, mfaSecret); err != nil {
		return nil, err
	}

	return &MfaVerifyResult{Verified: true}, nil
}

func (s *MfaService) Disable(ctx context.Context, userID string, code string) (*MfaDisableResult, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return nil, utils.NewError(400, "MFA_NOT_ENABLED", "MFA is not enabled on this account.", nil)
	}
	if mfaSecret.ConfirmedAt == nil {
		return nil, utils.NewError(400, "MFA_NOT_ENABLED", "MFA is not enabled on this account.", nil)
	}

	valid := totp.Validate(code, mfaSecret.Secret)
	if !valid {
		return nil, utils.NewError(400, "INVALID_MFA_CODE", "The provided MFA code is invalid.", nil)
	}

	if err := s.repos.MfaSecrets().DeleteByUserID(ctx, userID); err != nil {
		return nil, err
	}
	_ = s.repos.MfaRecoveryCodes().DeleteByUserID(ctx, userID)

	return &MfaDisableResult{Disabled: true}, nil
}

func (s *MfaService) ValidateLogin(ctx context.Context, userID string, code string) (*MfaValidateLoginResult, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return &MfaValidateLoginResult{Valid: true}, nil
	}
	if mfaSecret.ConfirmedAt == nil {
		return &MfaValidateLoginResult{Valid: true}, nil
	}

	valid := totp.Validate(code, mfaSecret.Secret)
	if !valid {
		return nil, utils.NewError(400, "INVALID_MFA_CODE", "The provided MFA code is invalid.", nil)
	}

	return &MfaValidateLoginResult{Valid: true}, nil
}

func (s *MfaService) ValidateRecoveryCode(ctx context.Context, userID string, code string) (*MfaValidateLoginResult, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return &MfaValidateLoginResult{Valid: true}, nil
	}
	if mfaSecret.ConfirmedAt == nil {
		return &MfaValidateLoginResult{Valid: true}, nil
	}

	codes, err := s.repos.MfaRecoveryCodes().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	normalizedInput := strings.TrimSpace(strings.ToLower(code))
	codeHash := hashRecoveryCode(normalizedInput)

	for _, rc := range codes {
		if rc.UsedAt == nil && rc.CodeHash == codeHash {
			_ = s.repos.MfaRecoveryCodes().MarkUsed(ctx, rc.ID)
			return &MfaValidateLoginResult{Valid: true}, nil
		}
	}

	return nil, utils.NewError(400, "INVALID_RECOVERY_CODE", "The provided recovery code is invalid or already used.", nil)
}

func (s *MfaService) GetRecoveryCodes(ctx context.Context, userID string) (*MfaRecoveryCodesResult, error) {
	has, err := s.hasConfirmedMFA(ctx, userID)
	if err != nil || !has {
		return nil, utils.NewError(400, "MFA_NOT_ENABLED", "MFA is not enabled on this account.", nil)
	}

	codes, err := s.repos.MfaRecoveryCodes().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var remaining []string
	for _, rc := range codes {
		if rc.UsedAt == nil {
			remaining = append(remaining, rc.CodeHash)
		}
	}

	return &MfaRecoveryCodesResult{RecoveryCodes: remaining}, nil
}

func (s *MfaService) RegenerateRecoveryCodes(ctx context.Context, userID string, code string) (*MfaRecoveryCodesResult, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return nil, utils.NewError(400, "MFA_NOT_ENABLED", "MFA is not enabled on this account.", nil)
	}
	if mfaSecret.ConfirmedAt == nil {
		return nil, utils.NewError(400, "MFA_NOT_ENABLED", "MFA is not enabled on this account.", nil)
	}

	valid := totp.Validate(code, mfaSecret.Secret)
	if !valid {
		return nil, utils.NewError(400, "INVALID_MFA_CODE", "The provided MFA code is invalid.", nil)
	}

	_ = s.repos.MfaRecoveryCodes().DeleteByUserID(ctx, userID)

	recoveryCodes := s.generateRecoveryCodes(s.cfg.MFARecoveryCodeLength)
	if err := s.storeRecoveryCodes(ctx, userID, recoveryCodes); err != nil {
		return nil, err
	}

	return &MfaRecoveryCodesResult{RecoveryCodes: recoveryCodes}, nil
}

func (s *MfaService) HasMFA(ctx context.Context, userID string) bool {
	has, _ := s.hasConfirmedMFA(ctx, userID)
	return has
}

func (s *MfaService) hasConfirmedMFA(ctx context.Context, userID string) (bool, error) {
	mfaSecret, err := s.repos.MfaSecrets().GetByUserID(ctx, userID)
	if err != nil {
		return false, nil
	}
	return mfaSecret.ConfirmedAt != nil, nil
}

func (s *MfaService) generateRecoveryCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = generateOneTimeCode()
	}
	return codes
}

func generateOneTimeCode() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	hexStr := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s", hexStr[0:4], hexStr[4:8], hexStr[8:12])
}

func hashRecoveryCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func (s *MfaService) storeRecoveryCodes(ctx context.Context, userID string, codes []string) error {
	now := time.Now().UTC()
	batch := make([]*models.MfaRecoveryCode, 0, len(codes))
	for _, raw := range codes {
		batch = append(batch, &models.MfaRecoveryCode{
			ID:        utils.NewID(),
			CreatedAt: now,
			UserID:    userID,
			CodeHash:  hashRecoveryCode(strings.ToLower(strings.TrimSpace(raw))),
		})
	}
	return s.repos.MfaRecoveryCodes().CreateBatch(ctx, batch)
}
