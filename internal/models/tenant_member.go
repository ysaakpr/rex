package models

import (
	"time"

	"github.com/google/uuid"
)

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
	MemberStatusPending  MemberStatus = "pending"
)

type TenantMember struct {
	ID        uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID    `gorm:"type:uuid;not null" json:"tenant_id"`
	UserID    string       `gorm:"type:varchar(255);not null" json:"user_id"`
	RoleID    uuid.UUID    `gorm:"type:uuid;not null" json:"role_id"`
	Status    MemberStatus `gorm:"type:member_status;not null;default:'active'" json:"status"`
	InvitedBy *string      `gorm:"type:varchar(255)" json:"invited_by"`
	JoinedAt  time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`

	// Associations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Role   Role   `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (TenantMember) TableName() string {
	return "tenant_members"
}

type AddMemberInput struct {
	UserID string    `json:"user_id" binding:"required"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type UpdateMemberInput struct {
	RoleID *uuid.UUID    `json:"role_id,omitempty"`
	Status *MemberStatus `json:"status,omitempty"`
}

type MemberResponse struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenant_id"`
	UserID    string        `json:"user_id"`
	RoleID    uuid.UUID     `json:"role_id"`
	Role      *RoleResponse `json:"role,omitempty"`
	Status    MemberStatus  `json:"status"`
	InvitedBy *string       `json:"invited_by"`
	JoinedAt  time.Time     `json:"joined_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func (tm *TenantMember) ToResponse() *MemberResponse {
	resp := &MemberResponse{
		ID:        tm.ID,
		TenantID:  tm.TenantID,
		UserID:    tm.UserID,
		RoleID:    tm.RoleID,
		Status:    tm.Status,
		InvitedBy: tm.InvitedBy,
		JoinedAt:  tm.JoinedAt,
		CreatedAt: tm.CreatedAt,
		UpdatedAt: tm.UpdatedAt,
	}

	if tm.Role.ID != uuid.Nil {
		resp.Role = tm.Role.ToResponse()
	}

	return resp
}
