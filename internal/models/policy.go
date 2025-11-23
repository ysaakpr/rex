package models

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents a group of permissions (was Role)
type Policy struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	TenantID    *uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Associations
	Permissions []Permission `gorm:"many2many:policy_permissions;" json:"permissions,omitempty"`
	Roles       []Role       `gorm:"many2many:role_policies;foreignKey:ID;joinForeignKey:PolicyID;References:ID;joinReferences:RoleID;" json:"roles,omitempty"`
}

func (Policy) TableName() string {
	return "policies"
}

type CreatePolicyInput struct {
	Name        string     `json:"name" binding:"required,min=2,max=100"`
	Description string     `json:"description" binding:"omitempty,max=500"`
	TenantID    *uuid.UUID `json:"tenant_id"`
}

type UpdatePolicyInput struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
}

type AssignPermissionsInput struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required,min=1"`
}

type PolicyResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	TenantID    *uuid.UUID           `json:"tenant_id"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	Roles       []RoleResponse       `json:"roles,omitempty"`
	RolesCount  int                  `json:"roles_count"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (p *Policy) ToResponse() *PolicyResponse {
	resp := &PolicyResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		TenantID:    p.TenantID,
		IsSystem:    p.IsSystem,
		RolesCount:  len(p.Roles),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if len(p.Permissions) > 0 {
		resp.Permissions = make([]PermissionResponse, len(p.Permissions))
		for i, perm := range p.Permissions {
			resp.Permissions[i] = *perm.ToResponse()
		}
	}

	if len(p.Roles) > 0 {
		resp.Roles = make([]RoleResponse, len(p.Roles))
		for i, role := range p.Roles {
			resp.Roles[i] = *role.ToResponse()
		}
	}

	return resp
}
