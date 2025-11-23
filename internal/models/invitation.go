package models

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusExpired   InvitationStatus = "expired"
	InvitationStatusCancelled InvitationStatus = "cancelled"
)

type UserInvitation struct {
	ID         uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID        `gorm:"type:uuid;not null" json:"tenant_id"`
	Email      string           `gorm:"type:varchar(255);not null" json:"email"`
	InvitedBy  string           `gorm:"type:varchar(255);not null" json:"invited_by"`
	RoleID     uuid.UUID        `gorm:"type:uuid;not null" json:"role_id"`
	Token      string           `gorm:"type:varchar(255);unique;not null" json:"token"`
	Status     InvitationStatus `gorm:"type:invitation_status;not null;default:'pending'" json:"status"`
	AcceptedAt *time.Time       `json:"accepted_at"`
	ExpiresAt  time.Time        `gorm:"not null" json:"expires_at"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`

	// Associations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Role   Role   `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (UserInvitation) TableName() string {
	return "user_invitations"
}

type CreateInvitationInput struct {
	Email  string    `json:"email" binding:"required,email"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type InvitationResponse struct {
	ID         uuid.UUID        `json:"id"`
	TenantID   uuid.UUID        `json:"tenant_id"`
	Email      string           `json:"email"`
	InvitedBy  string           `json:"invited_by"`
	RoleID     uuid.UUID        `json:"role_id"`
	Role       *RoleResponse    `json:"role,omitempty"`
	Status     InvitationStatus `json:"status"`
	AcceptedAt *time.Time       `json:"accepted_at"`
	ExpiresAt  time.Time        `json:"expires_at"`
	CreatedAt  time.Time        `json:"created_at"`
}

func (inv *UserInvitation) ToResponse() *InvitationResponse {
	resp := &InvitationResponse{
		ID:         inv.ID,
		TenantID:   inv.TenantID,
		Email:      inv.Email,
		InvitedBy:  inv.InvitedBy,
		RoleID:     inv.RoleID,
		Status:     inv.Status,
		AcceptedAt: inv.AcceptedAt,
		ExpiresAt:  inv.ExpiresAt,
		CreatedAt:  inv.CreatedAt,
	}

	if inv.Role.ID != uuid.Nil {
		resp.Role = inv.Role.ToResponse()
	}

	return resp
}

func (inv *UserInvitation) IsExpired() bool {
	return time.Now().After(inv.ExpiresAt)
}

func (inv *UserInvitation) CanAccept() bool {
	return inv.Status == InvitationStatusPending && !inv.IsExpired()
}
