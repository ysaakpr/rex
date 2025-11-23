package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user's role in a tenant (was Relation)
// Examples: Admin, Writer, Viewer, Basic
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Type        string     `gorm:"type:varchar(20);not null" json:"type"` // tenant, platform, etc.
	Description string     `gorm:"type:text" json:"description"`
	TenantID    *uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Associations
	Policies []Policy `gorm:"many2many:role_policies;" json:"policies,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

type CreateRoleInput struct {
	Name        string     `json:"name" binding:"required,min=2,max=100"`
	Type        string     `json:"type" binding:"required,oneof=tenant platform"`
	Description string     `json:"description" binding:"omitempty,max=500"`
	TenantID    *uuid.UUID `json:"tenant_id"`
}

type UpdateRoleInput struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
}

type AssignPoliciesToRoleInput struct {
	PolicyIDs []uuid.UUID `json:"policy_ids" binding:"required,min=1"`
}

type RoleResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Description string           `json:"description"`
	TenantID    *uuid.UUID       `json:"tenant_id"`
	IsSystem    bool             `json:"is_system"`
	Policies    []PolicyResponse `json:"policies,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (r *Role) ToResponse() *RoleResponse {
	resp := &RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Type:        r.Type,
		Description: r.Description,
		TenantID:    r.TenantID,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if len(r.Policies) > 0 {
		resp.Policies = make([]PolicyResponse, len(r.Policies))
		for i, policy := range r.Policies {
			resp.Policies[i] = *policy.ToResponse()
		}
	}

	return resp
}
