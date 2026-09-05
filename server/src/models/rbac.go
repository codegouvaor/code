package models

import (
	"time"

	"gorm.io/datatypes"
)

type Role struct {
	Common
	Name        string         `gorm:"column:name;type:text;uniqueIndex;not null" json:"name"`
	Slug        string         `gorm:"column:slug;type:text;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	IsSystem    bool           `gorm:"column:is_system;default:false" json:"isSystem"`
	Permissions datatypes.JSON `gorm:"column:permissions;type:jsonb" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type UserRole struct {
	UserID     string    `gorm:"column:user_id;type:text;primaryKey" json:"userId"`
	RoleID     string    `gorm:"column:role_id;type:text;primaryKey" json:"roleId"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null" json:"assignedAt"`
}

func (UserRole) TableName() string { return "user_roles" }
