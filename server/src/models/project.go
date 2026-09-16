package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Project is the Code-native unit of the platform. It is deliberately not a
// GitHub/GitLab repository: a project owns a stable Code reference
// (for example "codegouvaor/react-ads") that survives provider changes, and
// zero or more RepositoryBinding rows that point at external forges.
type Project struct {
	Common
	Archivable
	OrganizationID  *string        `gorm:"column:organization_id;type:text;index" json:"organizationId,omitempty"`
	OwnerID         string         `gorm:"column:owner_id;type:text;index;not null" json:"ownerId"`
	Namespace       string         `gorm:"column:namespace;type:text;index;not null" json:"namespace"`
	Slug            string         `gorm:"column:slug;type:text;not null" json:"slug"`
	Reference       string         `gorm:"column:reference;type:text;uniqueIndex;not null" json:"reference"`
	Name            string         `gorm:"column:name;type:text;not null" json:"name"`
	Description     string         `gorm:"column:description;type:text" json:"description"`
	Visibility      string         `gorm:"column:visibility;type:text;not null;default:'public'" json:"visibility"`
	Topics          datatypes.JSON `gorm:"column:topics;type:jsonb" json:"topics"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	DefaultProvider string         `gorm:"column:default_provider;type:text" json:"defaultProvider,omitempty"`
	HomepageURL     *string        `gorm:"column:homepage_url;type:text" json:"homepageUrl,omitempty"`
	LastActivityAt  *time.Time     `gorm:"column:last_activity_at" json:"lastActivityAt,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Project) TableName() string { return "projects" }

type ProjectMember struct {
	Common
	ProjectID string    `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	UserID    string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Role      string    `gorm:"column:role;type:text;not null;default:'member'" json:"role"`
	JoinedAt  time.Time `gorm:"column:joined_at;not null" json:"joinedAt"`
	InvitedBy *string   `gorm:"column:invited_by;type:text" json:"invitedBy,omitempty"`
}

func (ProjectMember) TableName() string { return "project_members" }

// ProjectAssetKind enumerates the Code-native resources a project can own
// besides a repository binding (documentation, APIs, SDKs, services,
// standards, integrations).
type ProjectAssetKind string

const (
	ProjectAssetDocumentation ProjectAssetKind = "documentation"
	ProjectAssetAPI           ProjectAssetKind = "api"
	ProjectAssetSDK           ProjectAssetKind = "sdk"
	ProjectAssetService       ProjectAssetKind = "service"
	ProjectAssetStandard      ProjectAssetKind = "standard"
	ProjectAssetIntegration   ProjectAssetKind = "integration"
)

// ProjectAsset is the generic primitive used to attach Code-native resources
// to a project. Keeping a single table with a "kind" discriminator avoids
// freezing the domain into premature, feature-specific schemas while still
// giving each resource a stable identity, slug and metadata bag.
type ProjectAsset struct {
	Common
	ProjectID   string         `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	Kind        string         `gorm:"column:kind;type:text;index;not null" json:"kind"`
	Slug        string         `gorm:"column:slug;type:text;not null" json:"slug"`
	Name        string         `gorm:"column:name;type:text;not null" json:"name"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	URL         *string        `gorm:"column:url;type:text" json:"url,omitempty"`
	Visibility  string         `gorm:"column:visibility;type:text;not null;default:'public'" json:"visibility"`
	Position    int            `gorm:"column:position;not null;default:0" json:"position"`
	Metadata    datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
}

func (ProjectAsset) TableName() string { return "project_assets" }
