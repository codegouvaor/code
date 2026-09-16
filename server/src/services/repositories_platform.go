package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// topicJSON renders a single topic as a JSONB array literal so that it can be
// used with the PostgreSQL contains operator on the projects.topics column.
func topicJSON(topic string) string {
	encoded, err := json.Marshal([]string{topic})
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

// ── Accessors ────────────────────────────────────────────────────────────────

func (r *Repositories) Organizations() interfaces.OrganizationRepository {
	return &organizationRepository{db: r.db}
}

func (r *Repositories) OrganizationMembers() interfaces.OrganizationMemberRepository {
	return &organizationMemberRepository{db: r.db}
}

func (r *Repositories) Projects() interfaces.ProjectRepository {
	return &projectRepository{db: r.db}
}

func (r *Repositories) ProjectMembers() interfaces.ProjectMemberRepository {
	return &projectMemberRepository{db: r.db}
}

func (r *Repositories) ProjectAssets() interfaces.ProjectAssetRepository {
	return &projectAssetRepository{db: r.db}
}

func (r *Repositories) RepositoryBindings() interfaces.RepositoryBindingRepository {
	return &repositoryBindingRepository{db: r.db}
}

func (r *Repositories) ProviderConnections() interfaces.ProviderConnectionRepository {
	return &providerConnectionRepository{db: r.db}
}

func (r *Repositories) SyncCursors() interfaces.SyncCursorRepository {
	return &syncCursorRepository{db: r.db}
}

func (r *Repositories) WebhookSubscriptions() interfaces.WebhookSubscriptionRepository {
	return &webhookSubscriptionRepository{db: r.db}
}

func (r *Repositories) SyncJobs() interfaces.SyncJobRepository {
	return &syncJobRepository{db: r.db}
}

func (r *Repositories) ExternalResources() interfaces.ExternalResourceRepository {
	return &externalResourceRepository{db: r.db}
}

// ── Organizations ────────────────────────────────────────────────────────────

type organizationRepository struct{ db *gorm.DB }

func (r *organizationRepository) Create(ctx context.Context, organization *models.Organization) error {
	return r.db.WithContext(ctx).Create(organization).Error
}

func (r *organizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	var item models.Organization
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrOrganizationNotFound)
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var item models.Organization
	err := r.db.WithContext(ctx).First(&item, "slug = ?", strings.ToLower(strings.TrimSpace(slug))).Error
	return &item, normalizeNotFound(err, utils.ErrOrganizationNotFound)
}

func (r *organizationRepository) ListByUser(ctx context.Context, userID string) ([]models.Organization, error) {
	var items []models.Organization
	err := r.db.WithContext(ctx).
		Table("organizations").
		Joins("left join organization_members on organization_members.organization_id = organizations.id").
		Where("(organization_members.user_id = ? OR organizations.owner_id = ?) AND organizations.deleted_at IS NULL", userID, userID).
		Distinct("organizations.id, organizations.created_at, organizations.updated_at, organizations.slug, organizations.name, organizations.description, organizations.avatar_url, organizations.website_url, organizations.location, organizations.visibility, organizations.owner_id, organizations.metadata, organizations.archived_at, organizations.deleted_at").
		Order("organizations.name asc").
		Scan(&items).Error
	return items, err
}

func (r *organizationRepository) Search(ctx context.Context, query string, offset, limit int) ([]models.Organization, int64, error) {
	var items []models.Organization
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.Organization{}).Where("visibility = ?", "public")
	if query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		builder = builder.Where("lower(name) LIKE ? OR lower(slug) LIKE ? OR lower(description) LIKE ?", pattern, pattern, pattern)
	}
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("name asc").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *organizationRepository) Update(ctx context.Context, organization *models.Organization) error {
	return r.db.WithContext(ctx).Save(organization).Error
}

func (r *organizationRepository) Archive(ctx context.Context, id string, archivedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", id).Update("archived_at", archivedAt).Error
}

type organizationMemberRepository struct{ db *gorm.DB }

func (r *organizationMemberRepository) Create(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *organizationMemberRepository) Get(ctx context.Context, organizationID, userID string) (*models.OrganizationMember, error) {
	var item models.OrganizationMember
	err := r.db.WithContext(ctx).First(&item, "organization_id = ? AND user_id = ?", organizationID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrMembershipRequired)
}

func (r *organizationMemberRepository) ListByOrganization(ctx context.Context, organizationID string) ([]models.OrganizationMember, error) {
	var items []models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("joined_at asc").Find(&items).Error
	return items, err
}

func (r *organizationMemberRepository) Update(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *organizationMemberRepository) Delete(ctx context.Context, organizationID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.OrganizationMember{}, "organization_id = ? AND user_id = ?", organizationID, userID).Error
}

func (r *organizationMemberRepository) CountByOrganizationAndRole(ctx context.Context, organizationID, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND role = ?", organizationID, role).
		Count(&count).Error
	return count, err
}

// ── Projects ─────────────────────────────────────────────────────────────────

type projectRepository struct{ db *gorm.DB }

func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	var item models.Project
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrProjectNotFound)
}

func (r *projectRepository) GetByReference(ctx context.Context, reference string) (*models.Project, error) {
	var item models.Project
	err := r.db.WithContext(ctx).First(&item, "reference = ? OR slug = ?", reference, reference).Error
	return &item, normalizeNotFound(err, utils.ErrProjectNotFound)
}

func (r *projectRepository) GetByNamespaceAndSlug(ctx context.Context, namespace, slug string) (*models.Project, error) {
	var item models.Project
	err := r.db.WithContext(ctx).First(&item, "namespace = ? AND slug = ?", namespace, slug).Error
	return &item, normalizeNotFound(err, utils.ErrProjectNotFound)
}

func (r *projectRepository) ListByUser(ctx context.Context, userID string) ([]models.Project, error) {
	var items []models.Project
	err := r.db.WithContext(ctx).
		Table("projects").
		Joins("left join project_members on project_members.project_id = projects.id").
		Where("(project_members.user_id = ? OR projects.owner_id = ?) AND projects.deleted_at IS NULL", userID, userID).
		Distinct("projects.*").
		Order("projects.updated_at desc").
		Scan(&items).Error
	return items, err
}

func (r *projectRepository) ListByOrganization(ctx context.Context, organizationID string) ([]models.Project, error) {
	var items []models.Project
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("updated_at desc").
		Find(&items).Error
	return items, err
}

func (r *projectRepository) List(ctx context.Context, filter interfaces.ProjectFilter) ([]models.Project, int64, error) {
	var items []models.Project
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.Project{})
	if filter.OrganizationID != "" {
		builder = builder.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.OwnerID != "" {
		builder = builder.Where("owner_id = ?", filter.OwnerID)
	}
	if filter.Visibility != "" {
		builder = builder.Where("visibility = ?", filter.Visibility)
	}
	if filter.Topic != "" {
		builder = builder.Where("topics @> ?", topicJSON(filter.Topic))
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		builder = builder.Where("lower(name) LIKE ? OR lower(slug) LIKE ? OR lower(reference) LIKE ?", pattern, pattern, pattern)
	}
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("updated_at desc").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, count, err
}

func (r *projectRepository) Search(ctx context.Context, query string, offset, limit int) ([]models.Project, int64, error) {
	return r.projectSearch(ctx, query, offset, limit)
}

func (r *projectRepository) projectSearch(ctx context.Context, query string, offset, limit int) ([]models.Project, int64, error) {
	var items []models.Project
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.Project{}).Where("visibility = ?", "public")
	if query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		builder = builder.Where(
			"lower(name) LIKE ? OR lower(slug) LIKE ? OR lower(reference) LIKE ? OR lower(description) LIKE ? OR lower(topics::text) LIKE ?",
			pattern, pattern, pattern, pattern, pattern,
		)
	}
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("last_activity_at desc nulls last, updated_at desc").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *projectRepository) Archive(ctx context.Context, id string, archivedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Project{}).Where("id = ?", id).Update("archived_at", archivedAt).Error
}

type projectMemberRepository struct{ db *gorm.DB }

func (r *projectMemberRepository) Create(ctx context.Context, member *models.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *projectMemberRepository) Get(ctx context.Context, projectID, userID string) (*models.ProjectMember, error) {
	var item models.ProjectMember
	err := r.db.WithContext(ctx).First(&item, "project_id = ? AND user_id = ?", projectID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrMembershipRequired)
}

func (r *projectMemberRepository) ListByProject(ctx context.Context, projectID string) ([]models.ProjectMember, error) {
	var items []models.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("joined_at asc").Find(&items).Error
	return items, err
}

func (r *projectMemberRepository) Update(ctx context.Context, member *models.ProjectMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *projectMemberRepository) Delete(ctx context.Context, projectID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectMember{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}

type projectAssetRepository struct{ db *gorm.DB }

func (r *projectAssetRepository) Create(ctx context.Context, asset *models.ProjectAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *projectAssetRepository) GetByID(ctx context.Context, id string) (*models.ProjectAsset, error) {
	var item models.ProjectAsset
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrProjectAssetNotFound)
}

func (r *projectAssetRepository) GetByProjectAndSlug(ctx context.Context, projectID, kind, slug string) (*models.ProjectAsset, error) {
	var item models.ProjectAsset
	err := r.db.WithContext(ctx).First(&item, "project_id = ? AND kind = ? AND slug = ?", projectID, kind, slug).Error
	return &item, normalizeNotFound(err, utils.ErrProjectAssetNotFound)
}

func (r *projectAssetRepository) ListByProject(ctx context.Context, projectID, kind string) ([]models.ProjectAsset, error) {
	var items []models.ProjectAsset
	builder := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	if kind != "" {
		builder = builder.Where("kind = ?", kind)
	}
	err := builder.Order("position asc, name asc").Find(&items).Error
	return items, err
}

func (r *projectAssetRepository) Search(ctx context.Context, query string, offset, limit int) ([]models.ProjectAsset, int64, error) {
	var items []models.ProjectAsset
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.ProjectAsset{}).Where("visibility = ?", "public")
	if query != "" {
		pattern := "%" + strings.ToLower(query) + "%"
		builder = builder.Where("lower(name) LIKE ? OR lower(slug) LIKE ? OR lower(description) LIKE ?", pattern, pattern, pattern)
	}
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("name asc").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *projectAssetRepository) Update(ctx context.Context, asset *models.ProjectAsset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

func (r *projectAssetRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectAsset{}, "id = ?", id).Error
}

// ── Repository bindings ──────────────────────────────────────────────────────

type repositoryBindingRepository struct{ db *gorm.DB }

func (r *repositoryBindingRepository) Create(ctx context.Context, binding *models.RepositoryBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}

func (r *repositoryBindingRepository) GetByID(ctx context.Context, id string) (*models.RepositoryBinding, error) {
	var item models.RepositoryBinding
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrRepositoryBindingNotFound)
}

func (r *repositoryBindingRepository) GetByProjectAndProvider(ctx context.Context, projectID, provider string) (*models.RepositoryBinding, error) {
	var item models.RepositoryBinding
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND provider = ?", projectID, provider).
		Order("is_primary desc, created_at asc").
		First(&item).Error
	return &item, normalizeNotFound(err, utils.ErrRepositoryBindingNotFound)
}

func (r *repositoryBindingRepository) GetByProjectProviderAndExternalID(ctx context.Context, projectID, provider, externalID string) (*models.RepositoryBinding, error) {
	var item models.RepositoryBinding
	err := r.db.WithContext(ctx).
		First(&item, "project_id = ? AND provider = ? AND external_id = ?", projectID, provider, externalID).Error
	return &item, normalizeNotFound(err, utils.ErrRepositoryBindingNotFound)
}

func (r *repositoryBindingRepository) GetByExternal(ctx context.Context, owner, repo string) (*models.RepositoryBinding, error) {
	var item models.RepositoryBinding
	err := r.db.WithContext(ctx).
		Where("lower(external_owner) = ? AND lower(external_repo) = ?", strings.ToLower(owner), strings.ToLower(repo)).
		Order("is_primary desc, created_at asc").
		First(&item).Error
	return &item, normalizeNotFound(err, utils.ErrRepositoryBindingNotFound)
}

func (r *repositoryBindingRepository) ListByProject(ctx context.Context, projectID string) ([]models.RepositoryBinding, error) {
	var items []models.RepositoryBinding
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("is_primary desc, created_at asc").
		Find(&items).Error
	return items, err
}

func (r *repositoryBindingRepository) ListSyncable(ctx context.Context, staleBefore time.Time, limit int) ([]models.RepositoryBinding, error) {
	var items []models.RepositoryBinding
	err := r.db.WithContext(ctx).
		Where("sync_enabled = ? AND sync_status <> ? AND (last_synced_at IS NULL OR last_synced_at < ?)", true, models.BindingSyncDetached, staleBefore).
		Order("last_synced_at asc nulls first").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *repositoryBindingRepository) Update(ctx context.Context, binding *models.RepositoryBinding) error {
	return r.db.WithContext(ctx).Save(binding).Error
}

func (r *repositoryBindingRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.RepositoryBinding{}, "id = ?", id).Error
}

// ── Provider connections ─────────────────────────────────────────────────────

type providerConnectionRepository struct{ db *gorm.DB }

func (r *providerConnectionRepository) Create(ctx context.Context, connection *models.ProviderConnection) error {
	return r.db.WithContext(ctx).Create(connection).Error
}

func (r *providerConnectionRepository) GetByID(ctx context.Context, id string) (*models.ProviderConnection, error) {
	var item models.ProviderConnection
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrProviderConnectionNotFound)
}

func (r *providerConnectionRepository) GetByProviderAccount(ctx context.Context, provider, providerAccountID string) (*models.ProviderConnection, error) {
	var item models.ProviderConnection
	err := r.db.WithContext(ctx).First(&item, "provider = ? AND provider_account_id = ?", provider, providerAccountID).Error
	return &item, normalizeNotFound(err, utils.ErrProviderConnectionNotFound)
}

func (r *providerConnectionRepository) GetByUserAndProvider(ctx context.Context, userID, provider string) (*models.ProviderConnection, error) {
	var item models.ProviderConnection
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ? AND status = ?", userID, provider, models.ConnectionActive).
		Order("created_at desc").
		First(&item).Error
	return &item, normalizeNotFound(err, utils.ErrProviderConnectionNotFound)
}

func (r *providerConnectionRepository) ListByUser(ctx context.Context, userID string) ([]models.ProviderConnection, error) {
	var items []models.ProviderConnection
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at asc").Find(&items).Error
	return items, err
}

func (r *providerConnectionRepository) Update(ctx context.Context, connection *models.ProviderConnection) error {
	return r.db.WithContext(ctx).Save(connection).Error
}

func (r *providerConnectionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ProviderConnection{}, "id = ?", id).Error
}

// ── Sync cursors ─────────────────────────────────────────────────────────────

type syncCursorRepository struct{ db *gorm.DB }

func (r *syncCursorRepository) Get(ctx context.Context, bindingID, resource string) (*models.SyncCursor, error) {
	var item models.SyncCursor
	err := r.db.WithContext(ctx).First(&item, "binding_id = ? AND resource = ?", bindingID, resource).Error
	return &item, normalizeNotFound(err, utils.ErrSyncCursorNotFound)
}

func (r *syncCursorRepository) ListByBinding(ctx context.Context, bindingID string) ([]models.SyncCursor, error) {
	var items []models.SyncCursor
	err := r.db.WithContext(ctx).Where("binding_id = ?", bindingID).Order("resource asc").Find(&items).Error
	return items, err
}

func (r *syncCursorRepository) Upsert(ctx context.Context, cursor *models.SyncCursor) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "binding_id"}, {Name: "resource"}},
		UpdateAll: true,
	}).Create(cursor).Error
}

// ── Webhook subscriptions ────────────────────────────────────────────────────

type webhookSubscriptionRepository struct{ db *gorm.DB }

func (r *webhookSubscriptionRepository) Create(ctx context.Context, subscription *models.WebhookSubscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *webhookSubscriptionRepository) GetByID(ctx context.Context, id string) (*models.WebhookSubscription, error) {
	var item models.WebhookSubscription
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrWebhookSubscriptionNotFound)
}

func (r *webhookSubscriptionRepository) GetByBindingAndExternal(ctx context.Context, bindingID, externalID string) (*models.WebhookSubscription, error) {
	var item models.WebhookSubscription
	err := r.db.WithContext(ctx).First(&item, "binding_id = ? AND external_id = ?", bindingID, externalID).Error
	return &item, normalizeNotFound(err, utils.ErrWebhookSubscriptionNotFound)
}

func (r *webhookSubscriptionRepository) ListByBinding(ctx context.Context, bindingID string) ([]models.WebhookSubscription, error) {
	var items []models.WebhookSubscription
	err := r.db.WithContext(ctx).Where("binding_id = ?", bindingID).Order("created_at asc").Find(&items).Error
	return items, err
}

func (r *webhookSubscriptionRepository) ListActive(ctx context.Context, limit int) ([]models.WebhookSubscription, error) {
	var items []models.WebhookSubscription
	err := r.db.WithContext(ctx).Where("active = ?", true).Limit(limit).Find(&items).Error
	return items, err
}

func (r *webhookSubscriptionRepository) Update(ctx context.Context, subscription *models.WebhookSubscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *webhookSubscriptionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.WebhookSubscription{}, "id = ?", id).Error
}

// ── Sync jobs ────────────────────────────────────────────────────────────────

type syncJobRepository struct{ db *gorm.DB }

func (r *syncJobRepository) Create(ctx context.Context, job *models.SyncJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *syncJobRepository) GetByID(ctx context.Context, id string) (*models.SyncJob, error) {
	var item models.SyncJob
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrSyncJobNotFound)
}

func (r *syncJobRepository) GetByIdempotencyKey(ctx context.Context, key string) (*models.SyncJob, error) {
	var item models.SyncJob
	err := r.db.WithContext(ctx).First(&item, "idempotency_key = ?", key).Error
	return &item, normalizeNotFound(err, utils.ErrSyncJobNotFound)
}

func (r *syncJobRepository) ClaimNext(ctx context.Context, workerID string, now time.Time, lockFor time.Duration) (*models.SyncJob, error) {
	var claimed *models.SyncJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidate models.SyncJob
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status IN ? AND run_after <= ? AND (locked_until IS NULL OR locked_until < ?)", []string{models.SyncJobQueued, models.SyncJobFailed}, now, now).
			Order("run_after asc, created_at asc").
			First(&candidate).Error
		if err != nil {
			return err
		}
		lockedUntil := now.Add(lockFor)
		candidate.Status = models.SyncJobRunning
		candidate.LockedBy = workerID
		candidate.LockedUntil = &lockedUntil
		candidate.StartedAt = &now
		candidate.Attempts++
		candidate.UpdatedAt = now
		if err := tx.Save(&candidate).Error; err != nil {
			return err
		}
		claimed = &candidate
		return nil
	})
	if err != nil {
		return nil, normalizeNotFound(err, utils.ErrSyncJobNotFound)
	}
	return claimed, nil
}

func (r *syncJobRepository) Update(ctx context.Context, job *models.SyncJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *syncJobRepository) ListByStatus(ctx context.Context, status string, limit int) ([]models.SyncJob, error) {
	var items []models.SyncJob
	err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at desc").Limit(limit).Find(&items).Error
	return items, err
}

func (r *syncJobRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SyncJob{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// ── External resources ───────────────────────────────────────────────────────

type externalResourceRepository struct{ db *gorm.DB }

func (r *externalResourceRepository) Upsert(ctx context.Context, resource *models.ExternalResource) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "binding_id"}, {Name: "kind"}, {Name: "external_id"}},
		UpdateAll: true,
	}).Create(resource).Error
}

func (r *externalResourceRepository) GetByBindingAndExternal(ctx context.Context, bindingID, kind, externalID string) (*models.ExternalResource, error) {
	var item models.ExternalResource
	err := r.db.WithContext(ctx).First(&item, "binding_id = ? AND kind = ? AND external_id = ?", bindingID, kind, externalID).Error
	return &item, normalizeNotFound(err, utils.ErrExternalResourceNotFound)
}

func (r *externalResourceRepository) ListByProject(ctx context.Context, projectID, kind string, offset, limit int) ([]models.ExternalResource, int64, error) {
	var items []models.ExternalResource
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.ExternalResource{}).Where("project_id = ?", projectID)
	if kind != "" {
		builder = builder.Where("kind = ?", kind)
	}
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("external_updated_at desc nulls last, created_at desc").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *externalResourceRepository) DeleteByBinding(ctx context.Context, bindingID string) error {
	return r.db.WithContext(ctx).Delete(&models.ExternalResource{}, "binding_id = ?", bindingID).Error
}
