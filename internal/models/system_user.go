package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemUser represents a service account for machine-to-machine authentication
type SystemUser struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name            string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	ApplicationName string     `gorm:"type:varchar(100);not null;index" json:"application_name"` // Logical app name (multiple users can share)
	Email           string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	UserID          string     `gorm:"column:user_id;type:varchar(255);not null;uniqueIndex" json:"user_id"`
	Description     string     `gorm:"type:text" json:"description"`
	ServiceType     string     `gorm:"type:varchar(50);not null" json:"service_type"` // worker, integration, cron, api
	IsActive        bool       `gorm:"default:true;not null" json:"is_active"`
	IsPrimary       bool       `gorm:"default:true;not null;index" json:"is_primary"`   // Current/recommended credential
	ExpiresAt       *time.Time `gorm:"index" json:"expires_at"`                         // Grace period expiry (NULL = never expires)
	CreatedBy       string     `gorm:"type:varchar(255);not null" json:"created_by"`    // SuperTokens user ID of creator
	LastUsedAt      *time.Time `json:"last_used_at"`
	Metadata        JSONMap    `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (SystemUser) TableName() string {
	return "system_users"
}

// CreateSystemUserInput is the input for creating a system user
type CreateSystemUserInput struct {
	Name        string                 `json:"name" binding:"required,min=3,max=100"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
	ServiceType string                 `json:"service_type" binding:"required,oneof=worker integration cron api"`
	Metadata    map[string]interface{} `json:"metadata" binding:"omitempty"`
}

// UpdateSystemUserInput is the input for updating a system user
type UpdateSystemUserInput struct {
	Description *string                `json:"description,omitempty" binding:"omitempty,max=500"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SystemUserResponse is the response format for system user data
type SystemUserResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	ApplicationName string                 `json:"application_name"`
	Email           string                 `json:"email"`
	UserID          string                 `json:"user_id"`
	Description     string                 `json:"description"`
	ServiceType     string                 `json:"service_type"`
	IsActive        bool                   `json:"is_active"`
	IsPrimary       bool                   `json:"is_primary"`
	ExpiresAt       *time.Time             `json:"expires_at"`
	CreatedBy       string                 `json:"created_by"`
	LastUsedAt      *time.Time             `json:"last_used_at"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// SystemUserCreateResponse includes credentials that are only shown once
type SystemUserCreateResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	ApplicationName string                 `json:"application_name"`
	Email           string                 `json:"email"`
	Password        string                 `json:"password"` // Only included on creation
	UserID          string                 `json:"user_id"`
	Description     string                 `json:"description"`
	ServiceType     string                 `json:"service_type"`
	IsActive        bool                   `json:"is_active"`
	IsPrimary       bool                   `json:"is_primary"`
	ExpiresAt       *time.Time             `json:"expires_at"`
	CreatedBy       string                 `json:"created_by"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	Message         string                 `json:"message"`
	OldCredentials  []OldCredentialInfo    `json:"old_credentials,omitempty"` // During rotation
}

// OldCredentialInfo provides info about credentials being replaced
type OldCredentialInfo struct {
	Email     string     `json:"email"`
	ExpiresAt *time.Time `json:"expires_at"`
	Message   string     `json:"message"`
}

// ToResponse converts SystemUser to SystemUserResponse
func (s *SystemUser) ToResponse() *SystemUserResponse {
	metadata := make(map[string]interface{})
	if s.Metadata != nil {
		metadata = s.Metadata
	}

	return &SystemUserResponse{
		ID:              s.ID,
		Name:            s.Name,
		ApplicationName: s.ApplicationName,
		Email:           s.Email,
		UserID:          s.UserID,
		Description:     s.Description,
		ServiceType:     s.ServiceType,
		IsActive:        s.IsActive,
		IsPrimary:       s.IsPrimary,
		ExpiresAt:       s.ExpiresAt,
		CreatedBy:       s.CreatedBy,
		LastUsedAt:      s.LastUsedAt,
		Metadata:        metadata,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
