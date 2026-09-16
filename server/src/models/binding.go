package models

import (
	"time"

	"gorm.io/datatypes"
)

// Provider names supported by the platform layer. They are plain strings (and
// not a Go enum type) so that adding a provider never requires touching the
// domain models.
const (
	ProviderGitHub  = "github"
	ProviderGitLab  = "gitlab"
	ProviderGiteria = "giteria"
)

// Synchronisation states for a repository binding.
const (
	BindingSyncPending  = "pending"
	BindingSyncSyncing  = "syncing"
	BindingSyncSynced   = "synced"
	BindingSyncError    = "error"
	BindingSyncDetached = "detached"
)

// RepositoryBinding links a Code project to a repository hosted on an external
// forge. A project can have several bindings (a mirror on GitHub and a
// canonical repository on Giteria for instance), but the Code project keeps a
// single stable identity.
type RepositoryBinding struct {
	Common
	ProjectID        string         `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	Provider         string         `gorm:"column:provider;type:text;index;not null" json:"provider"`
	ExternalID       string         `gorm:"column:external_id;type:text;not null" json:"externalId"`
	ExternalURL      string         `gorm:"column:external_url;type:text" json:"externalUrl"`
	ExternalOwner    string         `gorm:"column:external_owner;type:text;not null" json:"externalOwner"`
	ExternalRepo     string         `gorm:"column:external_repo;type:text;not null" json:"externalRepo"`
	DefaultBranch    string         `gorm:"column:default_branch;type:text" json:"defaultBranch"`
	ConnectionID     *string        `gorm:"column:connection_id;type:text;index" json:"connectionId,omitempty"`
	IsPrimary        bool           `gorm:"column:is_primary;not null;default:false" json:"isPrimary"`
	SyncEnabled      bool           `gorm:"column:sync_enabled;not null;default:true" json:"syncEnabled"`
	SyncStatus       string         `gorm:"column:sync_status;type:text;not null;default:'pending'" json:"syncStatus"`
	LastSyncError    string         `gorm:"column:last_sync_error;type:text" json:"lastSyncError,omitempty"`
	LastSyncedAt     *time.Time     `gorm:"column:last_synced_at" json:"lastSyncedAt,omitempty"`
	ExternalUpdated  *time.Time     `gorm:"column:external_updated_at" json:"externalUpdatedAt,omitempty"`
	SyncMetadata     datatypes.JSON `gorm:"column:sync_metadata;type:jsonb" json:"syncMetadata,omitempty"`
	DetachedAt       *time.Time     `gorm:"column:detached_at" json:"detachedAt,omitempty"`
	ExternalArchived bool           `gorm:"column:external_archived;not null;default:false" json:"externalArchived"`
}

func (RepositoryBinding) TableName() string { return "repository_bindings" }

// Provider connection statuses.
const (
	ConnectionActive  = "active"
	ConnectionExpired = "expired"
	ConnectionRevoked = "revoked"
)

// ProviderConnection is a repository-scoped grant obtained through the
// existing OAuth flow. It is intentionally distinct from AuthAccount: an
// AuthAccount proves the Code identity (login), while a ProviderConnection
// carries the credentials used to read forge resources. Tokens are stored
// encrypted (never in clear text) and the scopes are limited to what the
// platform actually needs.
type ProviderConnection struct {
	Common
	UserID            string     `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Provider          string     `gorm:"column:provider;type:text;not null" json:"provider"`
	ProviderAccountID string     `gorm:"column:provider_account_id;type:text;not null" json:"providerAccountId"`
	AccountLogin      string     `gorm:"column:account_login;type:text" json:"accountLogin"`
	AvatarURL         *string    `gorm:"column:avatar_url;type:text" json:"avatarUrl,omitempty"`
	Scopes            string     `gorm:"column:scopes;type:text" json:"scopes"`
	AccessTokenEnc    string     `gorm:"column:access_token_enc;type:text" json:"-"`
	RefreshTokenEnc   string     `gorm:"column:refresh_token_enc;type:text" json:"-"`
	TokenExpiresAt    *time.Time `gorm:"column:token_expires_at" json:"tokenExpiresAt,omitempty"`
	Status            string     `gorm:"column:status;type:text;not null;default:'active'" json:"status"`
	LastValidatedAt   *time.Time `gorm:"column:last_validated_at" json:"lastValidatedAt,omitempty"`
	RevokedAt         *time.Time `gorm:"column:revoked_at" json:"revokedAt,omitempty"`
}

func (ProviderConnection) TableName() string { return "provider_connections" }

// SyncCursor stores the incremental position reached while synchronising one
// (binding, resource) pair. It is what makes synchronisation idempotent and
// replayable: events older than the cursor are ignored.
type SyncCursor struct {
	Common
	BindingID     string     `gorm:"column:binding_id;type:text;not null" json:"bindingId"`
	Provider      string     `gorm:"column:provider;type:text;not null" json:"provider"`
	Resource      string     `gorm:"column:resource;type:text;not null" json:"resource"`
	Cursor        string     `gorm:"column:cursor;type:text" json:"cursor"`
	LastEventID   string     `gorm:"column:last_event_id;type:text" json:"lastEventId"`
	LastEventAt   *time.Time `gorm:"column:last_event_at" json:"lastEventAt,omitempty"`
	LastEventKind string     `gorm:"column:last_event_kind;type:text" json:"lastEventKind"`
	Checksum      string     `gorm:"column:checksum;type:text" json:"checksum"`
}

func (SyncCursor) TableName() string { return "sync_cursors" }

// WebhookSubscription is the platform-side record of a webhook registered on
// the provider for a binding (secret, subscribed events, health).
type WebhookSubscription struct {
	Common
	BindingID      string         `gorm:"column:binding_id;type:text;index;not null" json:"bindingId"`
	Provider       string         `gorm:"column:provider;type:text;not null" json:"provider"`
	ExternalID     string         `gorm:"column:external_id;type:text;not null" json:"externalId"`
	Events         datatypes.JSON `gorm:"column:events;type:jsonb" json:"events"`
	SecretEnc      string         `gorm:"column:secret_enc;type:text" json:"-"`
	Active         bool           `gorm:"column:active;not null;default:true" json:"active"`
	LastDeliveryID string         `gorm:"column:last_delivery_id;type:text" json:"lastDeliveryId,omitempty"`
	LastReceivedAt *time.Time     `gorm:"column:last_received_at" json:"lastReceivedAt,omitempty"`
	FailureCount   int            `gorm:"column:failure_count;not null;default:0" json:"failureCount"`
}

func (WebhookSubscription) TableName() string { return "webhook_subscriptions" }

// Sync job kinds.
const (
	SyncJobRepositoryImport = "repository.import"
	SyncJobRepositorySync   = "repository.sync"
	SyncJobWebhookProcess   = "webhook.process"
	SyncJobSearchIndex      = "search.index"
	SyncJobDocIndex         = "documentation.index"
	SyncJobReconciliation   = "reconciliation"
)

// Sync job statuses.
const (
	SyncJobQueued    = "queued"
	SyncJobRunning   = "running"
	SyncJobSucceeded = "succeeded"
	SyncJobFailed    = "failed"
	SyncJobDead      = "dead"
)

// SyncJob is the durable record of an asynchronous platform operation. The
// idempotency key is unique so that duplicated webhook deliveries or retried
// requests collapse into a single job.
type SyncJob struct {
	Common
	Kind           string         `gorm:"column:kind;type:text;index;not null" json:"kind"`
	ProjectID      *string        `gorm:"column:project_id;type:text;index" json:"projectId,omitempty"`
	BindingID      *string        `gorm:"column:binding_id;type:text;index" json:"bindingId,omitempty"`
	Status         string         `gorm:"column:status;type:text;index;not null;default:'queued'" json:"status"`
	Attempts       int            `gorm:"column:attempts;not null;default:0" json:"attempts"`
	MaxAttempts    int            `gorm:"column:max_attempts;not null;default:5" json:"maxAttempts"`
	RunAfter       time.Time      `gorm:"column:run_after;not null;index" json:"runAfter"`
	StartedAt      *time.Time     `gorm:"column:started_at" json:"startedAt,omitempty"`
	FinishedAt     *time.Time     `gorm:"column:finished_at" json:"finishedAt,omitempty"`
	LastError      string         `gorm:"column:last_error;type:text" json:"lastError,omitempty"`
	IdempotencyKey string         `gorm:"column:idempotency_key;type:text;uniqueIndex;not null" json:"idempotencyKey"`
	Payload        datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	LockedBy       string         `gorm:"column:locked_by;type:text" json:"lockedBy,omitempty"`
	LockedUntil    *time.Time     `gorm:"column:locked_until" json:"lockedUntil,omitempty"`
}

func (SyncJob) TableName() string { return "sync_jobs" }

// External resource kinds stored in the local read model.
const (
	ExternalResourceIssue   = "issue"
	ExternalResourceReview  = "review"
	ExternalResourceRelease = "release"
)

// ExternalResource is the local read model of a provider resource. It lets the
// platform answer project/repository queries, feed search and expose stable
// identifiers even when the provider is temporarily unavailable.
type ExternalResource struct {
	Common
	BindingID         string         `gorm:"column:binding_id;type:text;index;not null" json:"bindingId"`
	ProjectID         string         `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	Provider          string         `gorm:"column:provider;type:text;not null" json:"provider"`
	Kind              string         `gorm:"column:kind;type:text;index;not null" json:"kind"`
	ExternalID        string         `gorm:"column:external_id;type:text;not null" json:"externalId"`
	Number            int            `gorm:"column:number;not null;default:0" json:"number"`
	Title             string         `gorm:"column:title;type:text" json:"title"`
	State             string         `gorm:"column:state;type:text;index" json:"state"`
	URL               string         `gorm:"column:url;type:text" json:"url"`
	AuthorLogin       string         `gorm:"column:author_login;type:text" json:"authorLogin"`
	ExternalUpdatedAt *time.Time     `gorm:"column:external_updated_at" json:"externalUpdatedAt,omitempty"`
	Payload           datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	SyncedAt          time.Time      `gorm:"column:synced_at;not null" json:"syncedAt"`
}

func (ExternalResource) TableName() string { return "external_resources" }
