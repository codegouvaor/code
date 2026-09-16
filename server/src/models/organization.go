package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Organization is a Code-native organisation. It is intentionally decoupled
// from Workspace: a workspace is a collaboration container inherited from the
// historical product, while an organization is the public/Code identity that
// owns projects, teams and provider integrations.
type Organization struct {
	Common
	Archivable
	Slug        string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"column:name;type:text;not null" json:"name"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	AvatarURL   *string        `gorm:"column:avatar_url;type:text" json:"avatarUrl,omitempty"`
	WebsiteURL  *string        `gorm:"column:website_url;type:text" json:"websiteUrl,omitempty"`
	Location    string         `gorm:"column:location;type:text" json:"location"`
	Visibility  string         `gorm:"column:visibility;type:text;not null;default:'public'" json:"visibility"`
	OwnerID     string         `gorm:"column:owner_id;type:text;index;not null" json:"ownerId"`
	Metadata    datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Organization) TableName() string { return "organizations" }

type OrganizationMember struct {
	Common
	OrganizationID string     `gorm:"column:organization_id;type:text;index;not null" json:"organizationId"`
	UserID         string     `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Role           string     `gorm:"column:role;type:text;not null;default:'member'" json:"role"`
	JoinedAt       time.Time  `gorm:"column:joined_at;not null" json:"joinedAt"`
	InvitedBy      *string    `gorm:"column:invited_by;type:text" json:"invitedBy,omitempty"`
	LastSeenAt     *time.Time `gorm:"column:last_seen_at" json:"lastSeenAt,omitempty"`
}

func (OrganizationMember) TableName() string { return "organization_members" }
