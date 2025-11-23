package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantStatus string

const (
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Status    TenantStatus   `gorm:"type:tenant_status;not null;default:'pending'" json:"status"`
	Metadata  JSONMap        `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedBy string         `gorm:"type:varchar(255);not null" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Tenant) TableName() string {
	return "tenants"
}

type CreateTenantInput struct {
	Name     string  `json:"name" binding:"required,min=3,max=255"`
	Slug     string  `json:"slug" binding:"required,min=3,max=255,alphanum"`
	Metadata JSONMap `json:"metadata"`
}

type UpdateTenantInput struct {
	Name     *string       `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Status   *TenantStatus `json:"status,omitempty"`
	Metadata JSONMap       `json:"metadata,omitempty"`
}

type TenantResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Status      TenantStatus `json:"status"`
	Metadata    JSONMap      `json:"metadata"`
	CreatedBy   string       `json:"created_by"`
	MemberCount int          `json:"member_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (t *Tenant) ToResponse() *TenantResponse {
	return &TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		Status:    t.Status,
		Metadata:  t.Metadata,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
