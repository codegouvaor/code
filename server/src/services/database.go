package services

import (
	"context"

	"github.com/codegouvaor/code/server/src/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseService struct {
	db *gorm.DB
}

func NewDatabaseService(dsn string) (*DatabaseService, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DatabaseService{db: db}, nil
}

func (d *DatabaseService) Gorm() *gorm.DB {
	return d.db
}

func (d *DatabaseService) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *DatabaseService) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *DatabaseService) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(fn)
}

func (d *DatabaseService) AutoMigrate() error {
	return d.db.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
		&models.NotificationPreference{},
		&models.LocalCredential{},
		&models.AuthSession{},
		&models.AuthRefreshToken{},
		&models.EmailVerificationToken{},
		&models.PasswordResetToken{},
		&models.AuthAuditEvent{},
		&models.AuthAccount{},
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.WorkspaceSSOConfig{},
		&models.Role{},
		&models.UserRole{},
		&models.MfaSecret{},
		&models.MfaRecoveryCode{},
		&models.Article{},
		&models.ArticleTag{},
		&models.Category{},
		&models.Tag{},
		&models.Media{},
		&models.Webhook{},
		&models.WebhookDelivery{},
		&models.SeoConfig{},
		&models.NewsletterSubscriber{},
		&models.Schedule{},
		// Platform layer (Code-native projects, providers and synchronisation).
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationTeam{},
		&models.OrganizationTeamMember{},
		&models.Project{},
		&models.ProjectMember{},
		&models.ProjectAsset{},
		&models.ProjectStar{},
		&models.ProjectWatch{},
		&models.RepositoryBinding{},
		&models.ProviderConnection{},
		&models.SyncCursor{},
		&models.WebhookSubscription{},
		&models.SyncJob{},
		&models.ExternalResource{},
	)
}
