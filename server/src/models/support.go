package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FAQ struct {
	Common
	Question  string `gorm:"column:question;type:text;not null" json:"question"`
	Answer    string `gorm:"column:answer;type:text;not null" json:"answer"`
	Category  string `gorm:"column:category;type:text;not null;default:'general'" json:"category"`
	SortOrder int    `gorm:"column:sort_order;default:0" json:"sortOrder"`
	IsActive  bool   `gorm:"column:is_active;default:true" json:"isActive"`
}

func (FAQ) TableName() string { return "faqs" }

type Ticket struct {
	Common
	UserID       string         `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Subject      string         `gorm:"column:subject;type:text;not null" json:"subject"`
	Description  string         `gorm:"column:description;type:text;not null" json:"description"`
	Status       string         `gorm:"column:status;type:text;not null;default:'open'" json:"status"`
	Priority     string         `gorm:"column:priority;type:text;not null;default:'medium'" json:"priority"`
	Category     string         `gorm:"column:category;type:text;not null;default:'general'" json:"category"`
	AssignedTo   *string        `gorm:"column:assigned_to;type:text" json:"assignedTo,omitempty"`
	ResolvedAt   *time.Time     `gorm:"column:resolved_at" json:"resolvedAt,omitempty"`
	ClosedAt     *time.Time     `gorm:"column:closed_at" json:"closedAt,omitempty"`
	Metadata     datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Ticket) TableName() string { return "tickets" }

type TicketReply struct {
	Common
	TicketID string `gorm:"column:ticket_id;type:text;index;not null" json:"ticketId"`
	UserID   string `gorm:"column:user_id;type:text;index;not null" json:"userId"`
	Content  string `gorm:"column:content;type:text;not null" json:"content"`
	IsStaff  bool   `gorm:"column:is_staff;default:false" json:"isStaff"`
}

func (TicketReply) TableName() string { return "ticket_replies" }

type ContactMessage struct {
	Common
	Name    string `gorm:"column:name;type:text;not null" json:"name"`
	Email   string `gorm:"column:email;type:text;not null" json:"email"`
	Subject string `gorm:"column:subject;type:text;not null" json:"subject"`
	Message string `gorm:"column:message;type:text;not null" json:"message"`
	Status  string `gorm:"column:status;type:text;not null;default:'unread'" json:"status"`
	ReadAt  *time.Time `gorm:"column:read_at" json:"readAt,omitempty"`
	RepliedAt *time.Time `gorm:"column:replied_at" json:"repliedAt,omitempty"`
}

func (ContactMessage) TableName() string { return "contact_messages" }
