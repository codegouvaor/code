package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Workspace struct {
	Common
	Archivable
	Name        string         `gorm:"column:name;type:text;not null" json:"name"`
	Slug        string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Visibility  string         `gorm:"column:visibility;type:text;not null;default:'private'" json:"visibility"`
	OwnerID     string         `gorm:"column:owner_id;type:text;index;not null" json:"ownerId"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Workspace) TableName() string { return "workspaces" }

type WorkspaceMember struct {
	Common
	WorkspaceID string     `gorm:"column:workspace_id;type:text;index;not null" json:"workspaceId"`
	UserID      string     `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Role        string     `gorm:"column:role;type:text;not null;default:'member'" json:"role"`
	JoinedAt    time.Time  `gorm:"column:joined_at;not null" json:"joinedAt"`
	LastSeenAt  *time.Time `gorm:"column:last_seen_at" json:"lastSeenAt,omitempty"`
}

func (WorkspaceMember) TableName() string { return "workspace_members" }

type WorkspaceSSOConfig struct {
	Common
	WorkspaceId       string         `gorm:"column:workspace_id;type:text;uniqueIndex;not null" json:"workspaceId"`
	Enabled           bool           `gorm:"column:enabled;not null;default:false" json:"enabled"`
	Provider          string         `gorm:"column:provider;type:text" json:"provider"`
	EnforceSso        bool           `gorm:"column:enforce_sso;not null;default:false" json:"enforceSso"`
	AllowPasswordAuth  bool          `gorm:"column:allow_password_auth;not null;default:true" json:"allowPasswordAuth"`
	AutoProvision     bool           `gorm:"column:auto_provision;not null;default:false" json:"autoProvision"`
	AllowIdpInitiated bool          `gorm:"column:allow_idp_initiated;not null;default:false" json:"allowIdpInitiated"`
	DomainHint        *string        `gorm:"column:domain_hint;type:text" json:"domainHint,omitempty"`
	IssuerUrl         *string        `gorm:"column:issuer_url;type:text" json:"issuerUrl,omitempty"`
	SsoUrl            *string        `gorm:"column:sso_url;type:text" json:"ssoUrl,omitempty"`
	EntityId          *string        `gorm:"column:entity_id;type:text" json:"entityId,omitempty"`
	ClientId          string         `gorm:"column:client_id;type:text" json:"clientId"`
	ClientSecret      string         `gorm:"column:client_secret;type:text" json:"-"`
	Certificate       *string        `gorm:"column:certificate;type:text" json:"certificate,omitempty"`
	AllowedDomains    *string        `gorm:"column:allowed_domains;type:text" json:"allowedDomains,omitempty"`
	DefaultRole       *string        `gorm:"column:default_role;type:text" json:"defaultRole,omitempty"`
	AttributeMapping  datatypes.JSON `gorm:"column:attribute_mapping;type:jsonb" json:"attributeMapping,omitempty"`
}

func (WorkspaceSSOConfig) TableName() string { return "workspace_sso_configs" }
