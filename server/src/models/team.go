package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OrganizationTeam groups members of an organization and grants them a role on
// the projects attached to the team.
type OrganizationTeam struct {
	Common
	OrganizationID string         `gorm:"column:organization_id;type:text;index;not null" json:"organizationId"`
	Slug           string         `gorm:"column:slug;type:text;not null" json:"slug"`
	Name           string         `gorm:"column:name;type:text;not null" json:"name"`
	Description    string         `gorm:"column:description;type:text" json:"description"`
	Privacy        string         `gorm:"column:privacy;type:text;not null;default:'visible'" json:"privacy"`
	Permission     string         `gorm:"column:permission;type:text;not null;default:'read'" json:"permission"`
	Metadata       datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (OrganizationTeam) TableName() string { return "organization_teams" }

type OrganizationTeamMember struct {
	Common
	TeamID    string    `gorm:"column:team_id;type:text;index;not null" json:"teamId"`
	UserID    string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Role      string    `gorm:"column:role;type:text;not null;default:'member'" json:"role"`
	JoinedAt  time.Time `gorm:"column:joined_at;not null" json:"joinedAt"`
	InvitedBy *string   `gorm:"column:invited_by;type:text" json:"invitedBy,omitempty"`
}

func (OrganizationTeamMember) TableName() string { return "organization_team_members" }

// ProjectStar is the platform equivalent of a GitHub star: it is attached to
// the Code project (not to a provider repository) so that it survives a
// provider change.
type ProjectStar struct {
	Common
	ProjectID string    `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	UserID    string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	StarredAt time.Time `gorm:"column:starred_at;not null" json:"starredAt"`
}

func (ProjectStar) TableName() string { return "project_stars" }

// ProjectWatch is a user subscription to a project's activity.
type ProjectWatch struct {
	Common
	ProjectID string    `gorm:"column:project_id;type:text;index;not null" json:"projectId"`
	UserID    string    `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Level     string    `gorm:"column:level;type:text;not null;default:'participating'" json:"level"`
	WatchedAt time.Time `gorm:"column:watched_at;not null" json:"watchedAt"`
}

func (ProjectWatch) TableName() string { return "project_watches" }
