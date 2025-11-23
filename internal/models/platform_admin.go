package models

import (
	"time"

	"github.com/google/uuid"
)

type PlatformAdmin struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:varchar(255);unique;not null" json:"user_id"`
	CreatedBy string    `gorm:"type:varchar(255)" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PlatformAdmin) TableName() string {
	return "platform_admins"
}

type PlatformAdminResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pa *PlatformAdmin) ToResponse() *PlatformAdminResponse {
	return &PlatformAdminResponse{
		ID:        pa.ID,
		UserID:    pa.UserID,
		CreatedBy: pa.CreatedBy,
		CreatedAt: pa.CreatedAt,
		UpdatedAt: pa.UpdatedAt,
	}
}

type CreatePlatformAdminInput struct {
	UserID string `json:"user_id" binding:"required"`
}
