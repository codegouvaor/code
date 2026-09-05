package services

import (
	"context"
	"errors"
	"time"

	"github.com/codegouvaor/info/server/src/interfaces"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repositories struct {
	db *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{db: db}
}

func (r *Repositories) Users() interfaces.UserRepository { return &userRepository{db: r.db} }
func (r *Repositories) UserSettings() interfaces.UserSettingsRepository {
	return &userSettingsRepository{db: r.db}
}
func (r *Repositories) NotificationPreferences() interfaces.NotificationPreferenceRepository {
	return &notificationPreferenceRepository{db: r.db}
}
func (r *Repositories) LocalCredentials() interfaces.LocalCredentialRepository {
	return &localCredentialRepository{db: r.db}
}
func (r *Repositories) AuthSessions() interfaces.AuthSessionRepository {
	return &authSessionRepository{db: r.db}
}
func (r *Repositories) AuthRefreshTokens() interfaces.AuthRefreshTokenRepository {
	return &authRefreshTokenRepository{db: r.db}
}
func (r *Repositories) EmailVerificationTokens() interfaces.EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: r.db}
}
func (r *Repositories) PasswordResetTokens() interfaces.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: r.db}
}
func (r *Repositories) AuthAuditEvents() interfaces.AuthAuditEventRepository {
	return &authAuditEventRepository{db: r.db}
}
func (r *Repositories) AuthAccounts() interfaces.AuthAccountRepository {
	return &authAccountRepository{db: r.db}
}
func (r *Repositories) Workspaces() interfaces.WorkspaceRepository {
	return &workspaceRepository{db: r.db}
}
func (r *Repositories) WorkspaceMembers() interfaces.WorkspaceMemberRepository {
	return &workspaceMemberRepository{db: r.db}
}
func (r *Repositories) WorkspaceSSOConfigs() interfaces.WorkspaceSSOConfigRepository {
	return &workspaceSSOConfigRepository{db: r.db}
}
func (r *Repositories) Roles() interfaces.RoleRepository { return &roleRepository{db: r.db} }
func (r *Repositories) UserRoles() interfaces.UserRoleRepository {
	return &userRoleRepository{db: r.db}
}
func (r *Repositories) MfaRecoveryCodes() interfaces.MfaRecoveryCodeRepository {
	return &mfaRecoveryCodeRepository{db: r.db}
}
func (r *Repositories) MfaSecrets() interfaces.MfaSecretRepository {
	return &mfaSecretRepository{db: r.db}
}
func (r *Repositories) Articles() interfaces.ArticleRepository {
	return &articleRepository{db: r.db}
}
func (r *Repositories) Categories() interfaces.CategoryRepository {
	return &categoryRepository{db: r.db}
}
func (r *Repositories) Tags() interfaces.TagRepository {
	return &tagRepository{db: r.db}
}
func (r *Repositories) Media() interfaces.MediaRepository {
	return &mediaRepository{db: r.db}
}
func (r *Repositories) Webhooks() interfaces.WebhookRepository {
	return &webhookRepository{db: r.db}
}
func (r *Repositories) WebhookDeliveries() interfaces.WebhookDeliveryRepository {
	return &webhookDeliveryRepository{db: r.db}
}
func (r *Repositories) SeoConfigs() interfaces.SeoConfigRepository {
	return &seoConfigRepository{db: r.db}
}
func (r *Repositories) NewsletterSubscribers() interfaces.NewsletterSubscriberRepository {
	return &newsletterSubscriberRepository{db: r.db}
}
func (r *Repositories) Schedules() interfaces.ScheduleRepository {
	return &scheduleRepository{db: r.db}
}
func (r *Repositories) WithDB(db *gorm.DB) *Repositories { return &Repositories{db: db} }

type userRepository struct{ db *gorm.DB }

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, normalizeNotFound(err, utils.NewError(404, "USER_NOT_FOUND", "The requested user was not found.", nil))
}
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "email_normalized = ? OR email = ?", email, email).Error
	return &user, normalizeNotFound(err, utils.NewError(404, "USER_NOT_FOUND", "The requested user was not found.", nil))
}
func (r *userRepository) ListStale(ctx context.Context, before time.Time, limit int) ([]models.User, error) {
	var items []models.User
	err := r.db.WithContext(ctx).
		Limit(limit).
		Find(&items).Error
	return items, err
}
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

type userSettingsRepository struct{ db *gorm.DB }

func (r *userSettingsRepository) GetByUserID(ctx context.Context, userID string) (*models.UserSettings, error) {
	var item models.UserSettings
	err := r.db.WithContext(ctx).First(&item, "user_id = ?", userID).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "USER_SETTINGS_NOT_FOUND", "The requested user settings were not found.", nil))
}

func (r *userSettingsRepository) Upsert(ctx context.Context, settings *models.UserSettings) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(settings).Error
}

type notificationPreferenceRepository struct{ db *gorm.DB }

func (r *notificationPreferenceRepository) GetByUserID(ctx context.Context, userID string) (*models.NotificationPreference, error) {
	var item models.NotificationPreference
	err := r.db.WithContext(ctx).First(&item, "user_id = ?", userID).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "NOTIFICATION_PREFERENCES_NOT_FOUND", "The requested notification preferences were not found.", nil))
}

func (r *notificationPreferenceRepository) Upsert(ctx context.Context, preference *models.NotificationPreference) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(preference).Error
}

type localCredentialRepository struct{ db *gorm.DB }

func (r *localCredentialRepository) Create(ctx context.Context, credential *models.LocalCredential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

func (r *localCredentialRepository) GetByUserID(ctx context.Context, userID string) (*models.LocalCredential, error) {
	var credential models.LocalCredential
	err := r.db.WithContext(ctx).First(&credential, "user_id = ?", userID).Error
	return &credential, normalizeNotFound(err, utils.NewError(404, "LOCAL_CREDENTIAL_NOT_FOUND", "The requested credential was not found.", nil))
}

func (r *localCredentialRepository) Update(ctx context.Context, credential *models.LocalCredential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

type authSessionRepository struct{ db *gorm.DB }

func (r *authSessionRepository) Create(ctx context.Context, session *models.AuthSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *authSessionRepository) GetByID(ctx context.Context, id string) (*models.AuthSession, error) {
	var session models.AuthSession
	err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error
	return &session, normalizeNotFound(err, utils.NewError(404, "AUTH_SESSION_NOT_FOUND", "The requested session was not found.", nil))
}

func (r *authSessionRepository) ListActiveByUser(ctx context.Context, userID string, now time.Time) ([]models.AuthSession, error) {
	var items []models.AuthSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at desc").
		Find(&items).Error
	return items, err
}

func (r *authSessionRepository) ListByUser(ctx context.Context, userID string) ([]models.AuthSession, error) {
	var items []models.AuthSession
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&items).Error
	return items, err
}

func (r *authSessionRepository) Update(ctx context.Context, session *models.AuthSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *authSessionRepository) Revoke(ctx context.Context, id string, reason string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.AuthSession{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{
		"revoked_at":        revokedAt,
		"revocation_reason": reason,
		"updated_at":        revokedAt,
	}).Error
}

func (r *authSessionRepository) RevokeAllByUser(ctx context.Context, userID string, reason string, revokedAt time.Time, exceptSessionID string) error {
	query := r.db.WithContext(ctx).Model(&models.AuthSession{}).Where("user_id = ? AND revoked_at IS NULL", userID)
	if exceptSessionID != "" {
		query = query.Where("id <> ?", exceptSessionID)
	}
	return query.Updates(map[string]any{
		"revoked_at":        revokedAt,
		"revocation_reason": reason,
		"updated_at":        revokedAt,
	}).Error
}

func (r *authSessionRepository) RevokeFamily(ctx context.Context, familyID string, reason string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.AuthSession{}).Where("refresh_token_family_id = ? AND revoked_at IS NULL", familyID).Updates(map[string]any{
		"revoked_at":        revokedAt,
		"revocation_reason": reason,
		"updated_at":        revokedAt,
	}).Error
}

func (r *authSessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&models.AuthSession{}).Error
}

type authRefreshTokenRepository struct{ db *gorm.DB }

func (r *authRefreshTokenRepository) Create(ctx context.Context, token *models.AuthRefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.AuthRefreshToken, error) {
	var token models.AuthRefreshToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	return &token, normalizeNotFound(err, utils.NewError(404, "REFRESH_TOKEN_NOT_FOUND", "The requested refresh token was not found.", nil))
}

func (r *authRefreshTokenRepository) GetByID(ctx context.Context, id string) (*models.AuthRefreshToken, error) {
	var token models.AuthRefreshToken
	err := r.db.WithContext(ctx).First(&token, "id = ?", id).Error
	return &token, normalizeNotFound(err, utils.NewError(404, "REFRESH_TOKEN_NOT_FOUND", "The requested refresh token was not found.", nil))
}

func (r *authRefreshTokenRepository) Update(ctx context.Context, token *models.AuthRefreshToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *authRefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.AuthRefreshToken{}).Where("family_id = ? AND revoked_at IS NULL", familyID).Updates(map[string]any{
		"revoked_at": revokedAt,
		"updated_at": revokedAt,
	}).Error
}

func (r *authRefreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&models.AuthRefreshToken{}).Error
}

type emailVerificationTokenRepository struct{ db *gorm.DB }

func (r *emailVerificationTokenRepository) Create(ctx context.Context, token *models.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *emailVerificationTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	return &token, normalizeNotFound(err, utils.NewError(404, "EMAIL_VERIFICATION_TOKEN_NOT_FOUND", "The requested email verification token was not found.", nil))
}

func (r *emailVerificationTokenRepository) Update(ctx context.Context, token *models.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *emailVerificationTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&models.EmailVerificationToken{}).Error
}

type passwordResetTokenRepository struct{ db *gorm.DB }

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *passwordResetTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	return &token, normalizeNotFound(err, utils.NewError(404, "PASSWORD_RESET_TOKEN_NOT_FOUND", "The requested password reset token was not found.", nil))
}

func (r *passwordResetTokenRepository) Update(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *passwordResetTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&models.PasswordResetToken{}).Error
}

type authAuditEventRepository struct{ db *gorm.DB }

func (r *authAuditEventRepository) Create(ctx context.Context, event *models.AuthAuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

type workspaceRepository struct{ db *gorm.DB }

func (r *workspaceRepository) Create(ctx context.Context, workspace *models.Workspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}
func (r *workspaceRepository) ListByUser(ctx context.Context, userID string) ([]models.Workspace, error) {
	var items []models.Workspace
	err := r.db.WithContext(ctx).
		Table("workspaces").
		Joins("left join workspace_members on workspace_members.workspace_id = workspaces.id").
		Where("(workspace_members.user_id = ? OR workspaces.owner_id = ?) AND workspaces.archived_at IS NULL", userID, userID).
		Distinct("workspaces.id, workspaces.created_at, workspaces.updated_at, workspaces.name, workspaces.slug, workspaces.description, workspaces.visibility, workspaces.owner_id, workspaces.archived_at").
		Order("workspaces.created_at asc").
		Scan(&items).Error
	return items, err
}
func (r *workspaceRepository) GetByID(ctx context.Context, id string) (*models.Workspace, error) {
	var item models.Workspace
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrWorkspaceNotFound)
}
func (r *workspaceRepository) Update(ctx context.Context, workspace *models.Workspace) error {
	return r.db.WithContext(ctx).Save(workspace).Error
}
func (r *workspaceRepository) Archive(ctx context.Context, id string, archivedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Workspace{}).Where("id = ?", id).Update("archived_at", archivedAt).Error
}

type workspaceMemberRepository struct{ db *gorm.DB }

func (r *workspaceMemberRepository) Create(ctx context.Context, member *models.WorkspaceMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}
func (r *workspaceMemberRepository) Get(ctx context.Context, workspaceID, userID string) (*models.WorkspaceMember, error) {
	var item models.WorkspaceMember
	err := r.db.WithContext(ctx).First(&item, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrMembershipRequired)
}
func (r *workspaceMemberRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]models.WorkspaceMember, error) {
	var items []models.WorkspaceMember
	err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("joined_at asc").Find(&items).Error
	return items, err
}
func (r *workspaceMemberRepository) Update(ctx context.Context, member *models.WorkspaceMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}
func (r *workspaceMemberRepository) Delete(ctx context.Context, workspaceID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.WorkspaceMember{}, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error
}

type authAccountRepository struct{ db *gorm.DB }

func (r *authAccountRepository) Create(ctx context.Context, account *models.AuthAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *authAccountRepository) GetByProvider(ctx context.Context, provider, providerAccountID string) (*models.AuthAccount, error) {
	var account models.AuthAccount
	err := r.db.WithContext(ctx).Where("provider = ? AND provider_account_id = ?", provider, providerAccountID).First(&account).Error
	return &account, normalizeNotFound(err, utils.NewError(404, "OAUTH_ACCOUNT_NOT_FOUND", "The OAuth account was not found.", nil))
}

func (r *authAccountRepository) ListByUserID(ctx context.Context, userID string) ([]models.AuthAccount, error) {
	var items []models.AuthAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *authAccountRepository) GetByUserIDAndProvider(ctx context.Context, userID, provider string) (*models.AuthAccount, error) {
	var account models.AuthAccount
	err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).First(&account).Error
	return &account, normalizeNotFound(err, utils.NewError(404, "OAUTH_ACCOUNT_NOT_FOUND", "The OAuth account was not found.", nil))
}

func (r *authAccountRepository) Update(ctx context.Context, account *models.AuthAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *authAccountRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.AuthAccount{}, "id = ?", id).Error
}

type workspaceSSOConfigRepository struct{ db *gorm.DB }

func (r *workspaceSSOConfigRepository) GetByWorkspaceID(ctx context.Context, workspaceID string) (*models.WorkspaceSSOConfig, error) {
	var item models.WorkspaceSSOConfig
	err := r.db.WithContext(ctx).First(&item, "workspace_id = ?", workspaceID).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "WORKSPACE_SSO_CONFIG_NOT_FOUND", "The requested workspace SSO config was not found.", nil))
}

func (r *workspaceSSOConfigRepository) Upsert(ctx context.Context, config *models.WorkspaceSSOConfig) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "workspace_id"}},
		UpdateAll: true,
	}).Create(config).Error
}

type roleRepository struct{ db *gorm.DB }

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	var item models.Role
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "ROLE_NOT_FOUND", "The requested role was not found.", nil))
}

func (r *roleRepository) GetBySlug(ctx context.Context, slug string) (*models.Role, error) {
	var item models.Role
	err := r.db.WithContext(ctx).First(&item, "slug = ?", slug).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "ROLE_NOT_FOUND", "The requested role was not found.", nil))
}

func (r *roleRepository) List(ctx context.Context) ([]models.Role, error) {
	var items []models.Role
	err := r.db.WithContext(ctx).Order("name ASC").Find(&items).Error
	return items, err
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, "id = ?", id).Error
}

type userRoleRepository struct{ db *gorm.DB }

func (r *userRoleRepository) Assign(ctx context.Context, userRole *models.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *userRoleRepository) Remove(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).Delete(&models.UserRole{}, "user_id = ? AND role_id = ?", userID, roleID).Error
}

func (r *userRoleRepository) GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error) {
	var item models.UserRole
	err := r.db.WithContext(ctx).First(&item, "user_id = ? AND role_id = ?", userID, roleID).Error
	return &item, normalizeNotFound(err, utils.NewError(404, "USER_ROLE_NOT_FOUND", "The user does not have this role.", nil))
}

func (r *userRoleRepository) ListByUser(ctx context.Context, userID string) ([]models.UserRole, error) {
	var items []models.UserRole
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *userRoleRepository) ListByRole(ctx context.Context, roleID string) ([]models.UserRole, error) {
	var items []models.UserRole
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&items).Error
	return items, err
}

func (r *userRoleRepository) CountByRole(ctx context.Context, roleID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func normalizeNotFound(err error, notFound error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return err
}

// Article repository

type articleRepository struct{ db *gorm.DB }

func (r *articleRepository) Create(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *articleRepository) GetByID(ctx context.Context, id string) (*models.Article, error) {
	var article models.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").First(&article, "id = ?", id).Error
	return &article, normalizeNotFound(err, utils.NewError(404, "ARTICLE_NOT_FOUND", "The requested article was not found.", nil))
}

func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	var article models.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").First(&article, "slug = ?", slug).Error
	return &article, normalizeNotFound(err, utils.NewError(404, "ARTICLE_NOT_FOUND", "The requested article was not found.", nil))
}

func (r *articleRepository) List(ctx context.Context, workspaceID string, status string, categoryID string, offset, limit int) ([]models.Article, int64, error) {
	var articles []models.Article
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Article{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	query = query.Where("archived_at IS NULL")
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("Category").Preload("Author").Order("updated_at DESC").Offset(offset).Limit(limit).Find(&articles).Error
	return articles, count, err
}

func (r *articleRepository) Update(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Save(article).Error
}

func (r *articleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Article{}, "id = ?", id).Error
}

// Category repository

type categoryRepository struct{ db *gorm.DB }

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	return &category, normalizeNotFound(err, utils.NewError(404, "CATEGORY_NOT_FOUND", "The requested category was not found.", nil))
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, "slug = ?", slug).Error
	return &category, normalizeNotFound(err, utils.NewError(404, "CATEGORY_NOT_FOUND", "The requested category was not found.", nil))
}

func (r *categoryRepository) List(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).Where("archived_at IS NULL").Order("sort_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, "id = ?", id).Error
}

// Tag repository

type tagRepository struct{ db *gorm.DB }

func (r *tagRepository) Create(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error
	return &tag, normalizeNotFound(err, utils.NewError(404, "TAG_NOT_FOUND", "The requested tag was not found.", nil))
}

func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.WithContext(ctx).First(&tag, "slug = ?", slug).Error
	return &tag, normalizeNotFound(err, utils.NewError(404, "TAG_NOT_FOUND", "The requested tag was not found.", nil))
}

func (r *tagRepository) List(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.WithContext(ctx).Order("name ASC").Find(&tags).Error
	return tags, err
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Tag{}, "id = ?", id).Error
}

// Media repository

type mediaRepository struct{ db *gorm.DB }

func (r *mediaRepository) Create(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).First(&media, "id = ?", id).Error
	return &media, normalizeNotFound(err, utils.ErrMediaNotFound)
}

func (r *mediaRepository) List(ctx context.Context, workspaceID string, mimeType string, offset, limit int) ([]models.Media, int64, error) {
	var items []models.Media
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Media{})
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	if mimeType != "" {
		query = query.Where("mime_type LIKE ?", mimeType+"%")
	}
	query = query.Where("archived_at IS NULL")
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *mediaRepository) Update(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Media{}, "id = ?", id).Error
}

// Webhook repository

type webhookRepository struct{ db *gorm.DB }

func (r *webhookRepository) Create(ctx context.Context, webhook *models.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

func (r *webhookRepository) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	var webhook models.Webhook
	err := r.db.WithContext(ctx).First(&webhook, "id = ?", id).Error
	return &webhook, normalizeNotFound(err, utils.NewError(404, "WEBHOOK_NOT_FOUND", "The requested webhook was not found.", nil))
}

func (r *webhookRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]models.Webhook, error) {
	var items []models.Webhook
	err := r.db.WithContext(ctx).Where("workspace_id = ? AND archived_at IS NULL", workspaceID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *webhookRepository) Update(ctx context.Context, webhook *models.Webhook) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Webhook{}, "id = ?", id).Error
}

// WebhookDelivery repository

type webhookDeliveryRepository struct{ db *gorm.DB }

func (r *webhookDeliveryRepository) Create(ctx context.Context, delivery *models.WebhookDelivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *webhookDeliveryRepository) ListByWebhook(ctx context.Context, webhookID string, limit int) ([]models.WebhookDelivery, error) {
	var items []models.WebhookDelivery
	err := r.db.WithContext(ctx).Where("webhook_id = ?", webhookID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

// SeoConfig repository

type seoConfigRepository struct{ db *gorm.DB }

func (r *seoConfigRepository) GetByPagePath(ctx context.Context, pagePath string) (*models.SeoConfig, error) {
	var config models.SeoConfig
	err := r.db.WithContext(ctx).First(&config, "page_path = ?", pagePath).Error
	return &config, normalizeNotFound(err, utils.NewError(404, "SEO_CONFIG_NOT_FOUND", "The requested SEO config was not found.", nil))
}

func (r *seoConfigRepository) Upsert(ctx context.Context, config *models.SeoConfig) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "page_path"}},
		UpdateAll: true,
	}).Create(config).Error
}

func (r *seoConfigRepository) List(ctx context.Context) ([]models.SeoConfig, error) {
	var items []models.SeoConfig
	err := r.db.WithContext(ctx).Order("page_path ASC").Find(&items).Error
	return items, err
}

// NewsletterSubscriber repository

type newsletterSubscriberRepository struct{ db *gorm.DB }

func (r *newsletterSubscriberRepository) Create(ctx context.Context, subscriber *models.NewsletterSubscriber) error {
	return r.db.WithContext(ctx).Create(subscriber).Error
}

func (r *newsletterSubscriberRepository) GetByEmail(ctx context.Context, email string) (*models.NewsletterSubscriber, error) {
	var subscriber models.NewsletterSubscriber
	err := r.db.WithContext(ctx).First(&subscriber, "email_normalized = ?", email).Error
	return &subscriber, normalizeNotFound(err, utils.NewError(404, "SUBSCRIBER_NOT_FOUND", "The requested subscriber was not found.", nil))
}

func (r *newsletterSubscriberRepository) List(ctx context.Context, status string, offset, limit int) ([]models.NewsletterSubscriber, int64, error) {
	var items []models.NewsletterSubscriber
	var count int64
	query := r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("subscribed_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *newsletterSubscriberRepository) Update(ctx context.Context, subscriber *models.NewsletterSubscriber) error {
	return r.db.WithContext(ctx).Save(subscriber).Error
}

// Schedule repository

type scheduleRepository struct{ db *gorm.DB }

func (r *scheduleRepository) Create(ctx context.Context, schedule *models.Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepository) GetByID(ctx context.Context, id string) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.db.WithContext(ctx).First(&schedule, "id = ?", id).Error
	return &schedule, normalizeNotFound(err, utils.NewError(404, "SCHEDULE_NOT_FOUND", "The requested schedule was not found.", nil))
}

func (r *scheduleRepository) List(ctx context.Context, from, to time.Time, offset, limit int) ([]models.Schedule, int64, error) {
	var items []models.Schedule
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Schedule{})
	if !from.IsZero() {
		query = query.Where("scheduled_at >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("scheduled_at <= ?", to)
	}
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("scheduled_at ASC").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *scheduleRepository) Update(ctx context.Context, schedule *models.Schedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

func (r *scheduleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Schedule{}, "id = ?", id).Error
}
