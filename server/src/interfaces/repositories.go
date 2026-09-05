package interfaces

import (
	"context"
	"time"

	"github.com/codegouvaor/info/server/src/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ListStale(ctx context.Context, before time.Time, limit int) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type UserSettingsRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserSettings, error)
	Upsert(ctx context.Context, settings *models.UserSettings) error
}

type NotificationPreferenceRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.NotificationPreference, error)
	Upsert(ctx context.Context, preference *models.NotificationPreference) error
}

type LocalCredentialRepository interface {
	Create(ctx context.Context, credential *models.LocalCredential) error
	GetByUserID(ctx context.Context, userID string) (*models.LocalCredential, error)
	Update(ctx context.Context, credential *models.LocalCredential) error
}

type AuthSessionRepository interface {
	Create(ctx context.Context, session *models.AuthSession) error
	GetByID(ctx context.Context, id string) (*models.AuthSession, error)
	ListActiveByUser(ctx context.Context, userID string, now time.Time) ([]models.AuthSession, error)
	ListByUser(ctx context.Context, userID string) ([]models.AuthSession, error)
	Update(ctx context.Context, session *models.AuthSession) error
	Revoke(ctx context.Context, id string, reason string, revokedAt time.Time) error
	RevokeAllByUser(ctx context.Context, userID string, reason string, revokedAt time.Time, exceptSessionID string) error
	RevokeFamily(ctx context.Context, familyID string, reason string, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type AuthRefreshTokenRepository interface {
	Create(ctx context.Context, token *models.AuthRefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*models.AuthRefreshToken, error)
	GetByID(ctx context.Context, id string) (*models.AuthRefreshToken, error)
	Update(ctx context.Context, token *models.AuthRefreshToken) error
	RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type EmailVerificationTokenRepository interface {
	Create(ctx context.Context, token *models.EmailVerificationToken) error
	GetByHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error)
	Update(ctx context.Context, token *models.EmailVerificationToken) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *models.PasswordResetToken) error
	GetByHash(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error)
	Update(ctx context.Context, token *models.PasswordResetToken) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type AuthAuditEventRepository interface {
	Create(ctx context.Context, event *models.AuthAuditEvent) error
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *models.Workspace) error
	ListByUser(ctx context.Context, userID string) ([]models.Workspace, error)
	GetByID(ctx context.Context, id string) (*models.Workspace, error)
	Update(ctx context.Context, workspace *models.Workspace) error
	Archive(ctx context.Context, id string, archivedAt time.Time) error
}

type WorkspaceMemberRepository interface {
	Create(ctx context.Context, member *models.WorkspaceMember) error
	Get(ctx context.Context, workspaceID, userID string) (*models.WorkspaceMember, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]models.WorkspaceMember, error)
	Update(ctx context.Context, member *models.WorkspaceMember) error
	Delete(ctx context.Context, workspaceID, userID string) error
}

type AuthAccountRepository interface {
	Create(ctx context.Context, account *models.AuthAccount) error
	GetByProvider(ctx context.Context, provider string, providerAccountID string) (*models.AuthAccount, error)
	GetByUserIDAndProvider(ctx context.Context, userID string, provider string) (*models.AuthAccount, error)
	ListByUserID(ctx context.Context, userID string) ([]models.AuthAccount, error)
	Update(ctx context.Context, account *models.AuthAccount) error
	Delete(ctx context.Context, id string) error
}

type WorkspaceSSOConfigRepository interface {
	GetByWorkspaceID(ctx context.Context, workspaceID string) (*models.WorkspaceSSOConfig, error)
	Upsert(ctx context.Context, config *models.WorkspaceSSOConfig) error
}

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error)
	GetBySlug(ctx context.Context, slug string) (*models.Role, error)
	List(ctx context.Context) ([]models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id string) error
}

type UserRoleRepository interface {
	Assign(ctx context.Context, userRole *models.UserRole) error
	Remove(ctx context.Context, userID, roleID string) error
	GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error)
	ListByUser(ctx context.Context, userID string) ([]models.UserRole, error)
	ListByRole(ctx context.Context, roleID string) ([]models.UserRole, error)
	CountByRole(ctx context.Context, roleID string) (int64, error)
}

type MfaSecretRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.MfaSecret, error)
	Create(ctx context.Context, secret *models.MfaSecret) error
	Update(ctx context.Context, secret *models.MfaSecret) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type MfaRecoveryCodeRepository interface {
	Create(ctx context.Context, code *models.MfaRecoveryCode) error
	CreateBatch(ctx context.Context, codes []*models.MfaRecoveryCode) error
	GetByUserID(ctx context.Context, userID string) ([]models.MfaRecoveryCode, error)
	GetByID(ctx context.Context, id string) (*models.MfaRecoveryCode, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type ArticleRepository interface {
	Create(ctx context.Context, article *models.Article) error
	GetByID(ctx context.Context, id string) (*models.Article, error)
	GetBySlug(ctx context.Context, slug string) (*models.Article, error)
	List(ctx context.Context, workspaceID string, status string, categoryID string, offset, limit int) ([]models.Article, int64, error)
	Update(ctx context.Context, article *models.Article) error
	Delete(ctx context.Context, id string) error
}

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	List(ctx context.Context) ([]models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *models.Tag) error
	GetByID(ctx context.Context, id string) (*models.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tag, error)
	List(ctx context.Context) ([]models.Tag, error)
	Delete(ctx context.Context, id string) error
}

type MediaRepository interface {
	Create(ctx context.Context, media *models.Media) error
	GetByID(ctx context.Context, id string) (*models.Media, error)
	List(ctx context.Context, workspaceID string, mimeType string, offset, limit int) ([]models.Media, int64, error)
	Update(ctx context.Context, media *models.Media) error
	Delete(ctx context.Context, id string) error
}

type WebhookRepository interface {
	Create(ctx context.Context, webhook *models.Webhook) error
	GetByID(ctx context.Context, id string) (*models.Webhook, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]models.Webhook, error)
	Update(ctx context.Context, webhook *models.Webhook) error
	Delete(ctx context.Context, id string) error
}

type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *models.WebhookDelivery) error
	ListByWebhook(ctx context.Context, webhookID string, limit int) ([]models.WebhookDelivery, error)
}

type SeoConfigRepository interface {
	GetByPagePath(ctx context.Context, pagePath string) (*models.SeoConfig, error)
	Upsert(ctx context.Context, config *models.SeoConfig) error
	List(ctx context.Context) ([]models.SeoConfig, error)
}

type NewsletterSubscriberRepository interface {
	Create(ctx context.Context, subscriber *models.NewsletterSubscriber) error
	GetByEmail(ctx context.Context, email string) (*models.NewsletterSubscriber, error)
	List(ctx context.Context, status string, offset, limit int) ([]models.NewsletterSubscriber, int64, error)
	Update(ctx context.Context, subscriber *models.NewsletterSubscriber) error
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *models.Schedule) error
	GetByID(ctx context.Context, id string) (*models.Schedule, error)
	List(ctx context.Context, from, to time.Time, offset, limit int) ([]models.Schedule, int64, error)
	Update(ctx context.Context, schedule *models.Schedule) error
	Delete(ctx context.Context, id string) error
}

type RepositorySet interface {
	Users() UserRepository
	UserSettings() UserSettingsRepository
	NotificationPreferences() NotificationPreferenceRepository
	LocalCredentials() LocalCredentialRepository
	AuthSessions() AuthSessionRepository
	AuthRefreshTokens() AuthRefreshTokenRepository
	EmailVerificationTokens() EmailVerificationTokenRepository
	PasswordResetTokens() PasswordResetTokenRepository
	AuthAuditEvents() AuthAuditEventRepository
	AuthAccounts() AuthAccountRepository
	Workspaces() WorkspaceRepository
	WorkspaceMembers() WorkspaceMemberRepository
	WorkspaceSSOConfigs() WorkspaceSSOConfigRepository
	Roles() RoleRepository
	UserRoles() UserRoleRepository
	MfaSecrets() MfaSecretRepository
	MfaRecoveryCodes() MfaRecoveryCodeRepository
	Articles() ArticleRepository
	Categories() CategoryRepository
	Tags() TagRepository
	Media() MediaRepository
	Webhooks() WebhookRepository
	WebhookDeliveries() WebhookDeliveryRepository
	SeoConfigs() SeoConfigRepository
	NewsletterSubscribers() NewsletterSubscriberRepository
	Schedules() ScheduleRepository
}
