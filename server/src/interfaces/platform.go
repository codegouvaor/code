package interfaces

import (
	"context"
	"time"

	"github.com/codegouvaor/code/server/src/models"
)

// ── Organizations ────────────────────────────────────────────────────────────

type OrganizationRepository interface {
	Create(ctx context.Context, organization *models.Organization) error
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*models.Organization, error)
	ListByUser(ctx context.Context, userID string) ([]models.Organization, error)
	Search(ctx context.Context, query string, offset, limit int) ([]models.Organization, int64, error)
	Update(ctx context.Context, organization *models.Organization) error
	Archive(ctx context.Context, id string, archivedAt time.Time) error
}

type OrganizationTeamRepository interface {
	Create(ctx context.Context, team *models.OrganizationTeam) error
	GetByID(ctx context.Context, id string) (*models.OrganizationTeam, error)
	GetBySlug(ctx context.Context, organizationID, slug string) (*models.OrganizationTeam, error)
	ListByOrganization(ctx context.Context, organizationID string) ([]models.OrganizationTeam, error)
	Update(ctx context.Context, team *models.OrganizationTeam) error
	Delete(ctx context.Context, id string) error
}

type OrganizationTeamMemberRepository interface {
	Create(ctx context.Context, member *models.OrganizationTeamMember) error
	Get(ctx context.Context, teamID, userID string) (*models.OrganizationTeamMember, error)
	ListByTeam(ctx context.Context, teamID string) ([]models.OrganizationTeamMember, error)
	ListByUser(ctx context.Context, userID string) ([]models.OrganizationTeamMember, error)
	Delete(ctx context.Context, teamID, userID string) error
}

type ProjectStarRepository interface {
	Create(ctx context.Context, star *models.ProjectStar) error
	Delete(ctx context.Context, projectID, userID string) error
	Get(ctx context.Context, projectID, userID string) (*models.ProjectStar, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]models.ProjectStar, int64, error)
	CountByProject(ctx context.Context, projectID string) (int64, error)
}

type ProjectWatchRepository interface {
	Upsert(ctx context.Context, watch *models.ProjectWatch) error
	Delete(ctx context.Context, projectID, userID string) error
	Get(ctx context.Context, projectID, userID string) (*models.ProjectWatch, error)
	ListByUser(ctx context.Context, userID string) ([]models.ProjectWatch, error)
}

type OrganizationMemberRepository interface {
	Create(ctx context.Context, member *models.OrganizationMember) error
	Get(ctx context.Context, organizationID, userID string) (*models.OrganizationMember, error)
	ListByOrganization(ctx context.Context, organizationID string) ([]models.OrganizationMember, error)
	Update(ctx context.Context, member *models.OrganizationMember) error
	Delete(ctx context.Context, organizationID, userID string) error
	CountByOrganizationAndRole(ctx context.Context, organizationID, role string) (int64, error)
}

// ── Projects ─────────────────────────────────────────────────────────────────

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id string) (*models.Project, error)
	GetByReference(ctx context.Context, reference string) (*models.Project, error)
	GetByNamespaceAndSlug(ctx context.Context, namespace, slug string) (*models.Project, error)
	ListByUser(ctx context.Context, userID string) ([]models.Project, error)
	ListByOrganization(ctx context.Context, organizationID string) ([]models.Project, error)
	List(ctx context.Context, filter ProjectFilter) ([]models.Project, int64, error)
	Search(ctx context.Context, query string, offset, limit int) ([]models.Project, int64, error)
	Update(ctx context.Context, project *models.Project) error
	Archive(ctx context.Context, id string, archivedAt time.Time) error
}

// ProjectFilter narrows a project listing. Empty fields are ignored.
type ProjectFilter struct {
	OrganizationID string
	OwnerID        string
	Visibility     string
	Topic          string
	Query          string
	Offset         int
	Limit          int
}

type ProjectMemberRepository interface {
	Create(ctx context.Context, member *models.ProjectMember) error
	Get(ctx context.Context, projectID, userID string) (*models.ProjectMember, error)
	ListByProject(ctx context.Context, projectID string) ([]models.ProjectMember, error)
	Update(ctx context.Context, member *models.ProjectMember) error
	Delete(ctx context.Context, projectID, userID string) error
}

type ProjectAssetRepository interface {
	Create(ctx context.Context, asset *models.ProjectAsset) error
	GetByID(ctx context.Context, id string) (*models.ProjectAsset, error)
	GetByProjectAndSlug(ctx context.Context, projectID, kind, slug string) (*models.ProjectAsset, error)
	ListByProject(ctx context.Context, projectID, kind string) ([]models.ProjectAsset, error)
	Search(ctx context.Context, query string, offset, limit int) ([]models.ProjectAsset, int64, error)
	Update(ctx context.Context, asset *models.ProjectAsset) error
	Delete(ctx context.Context, id string) error
}

// ── Repository bindings ──────────────────────────────────────────────────────

type RepositoryBindingRepository interface {
	Create(ctx context.Context, binding *models.RepositoryBinding) error
	GetByID(ctx context.Context, id string) (*models.RepositoryBinding, error)
	GetByProjectAndProvider(ctx context.Context, projectID, provider string) (*models.RepositoryBinding, error)
	GetByProjectProviderAndExternalID(ctx context.Context, projectID, provider, externalID string) (*models.RepositoryBinding, error)
	GetByExternal(ctx context.Context, owner, repo string) (*models.RepositoryBinding, error)
	ListByProject(ctx context.Context, projectID string) ([]models.RepositoryBinding, error)
	ListSyncable(ctx context.Context, staleBefore time.Time, limit int) ([]models.RepositoryBinding, error)
	Update(ctx context.Context, binding *models.RepositoryBinding) error
	Delete(ctx context.Context, id string) error
}

type ProviderConnectionRepository interface {
	Create(ctx context.Context, connection *models.ProviderConnection) error
	GetByID(ctx context.Context, id string) (*models.ProviderConnection, error)
	GetByProviderAccount(ctx context.Context, provider, providerAccountID string) (*models.ProviderConnection, error)
	GetByUserAndProvider(ctx context.Context, userID, provider string) (*models.ProviderConnection, error)
	ListByUser(ctx context.Context, userID string) ([]models.ProviderConnection, error)
	Update(ctx context.Context, connection *models.ProviderConnection) error
	Delete(ctx context.Context, id string) error
}

// ── Synchronisation ──────────────────────────────────────────────────────────

type SyncCursorRepository interface {
	Get(ctx context.Context, bindingID, resource string) (*models.SyncCursor, error)
	ListByBinding(ctx context.Context, bindingID string) ([]models.SyncCursor, error)
	Upsert(ctx context.Context, cursor *models.SyncCursor) error
}

type WebhookSubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.WebhookSubscription) error
	GetByID(ctx context.Context, id string) (*models.WebhookSubscription, error)
	GetByBindingAndExternal(ctx context.Context, bindingID, externalID string) (*models.WebhookSubscription, error)
	ListByBinding(ctx context.Context, bindingID string) ([]models.WebhookSubscription, error)
	ListActive(ctx context.Context, limit int) ([]models.WebhookSubscription, error)
	Update(ctx context.Context, subscription *models.WebhookSubscription) error
	Delete(ctx context.Context, id string) error
}

type SyncJobRepository interface {
	Create(ctx context.Context, job *models.SyncJob) error
	GetByID(ctx context.Context, id string) (*models.SyncJob, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*models.SyncJob, error)
	ClaimNext(ctx context.Context, workerID string, now time.Time, lockFor time.Duration) (*models.SyncJob, error)
	Update(ctx context.Context, job *models.SyncJob) error
	ListByStatus(ctx context.Context, status string, limit int) ([]models.SyncJob, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

type ExternalResourceRepository interface {
	Upsert(ctx context.Context, resource *models.ExternalResource) error
	GetByBindingAndExternal(ctx context.Context, bindingID, kind, externalID string) (*models.ExternalResource, error)
	ListByProject(ctx context.Context, projectID, kind string, offset, limit int) ([]models.ExternalResource, int64, error)
	DeleteByBinding(ctx context.Context, bindingID string) error
}

// ── Search ───────────────────────────────────────────────────────────────────

// SearchFilter describes a platform-wide search request. It is deliberately
// provider- and engine-agnostic so that the PostgreSQL implementation can be
// replaced later without touching handlers.
type SearchFilter struct {
	Query   string
	Kinds   []string
	OwnerID string
	Offset  int
	Limit   int
}
