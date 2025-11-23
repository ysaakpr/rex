package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Service     string    `gorm:"type:varchar(100);not null" json:"service"`
	Entity      string    `gorm:"type:varchar(100);not null" json:"entity"`
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

type CreatePermissionInput struct {
	Service     string `json:"service" binding:"required,min=2,max=100"`
	Entity      string `json:"entity" binding:"required,min=2,max=100"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Service     string    `json:"service"`
	Entity      string    `json:"entity"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Key         string    `json:"key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Permission) ToResponse() *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Service:     p.Service,
		Entity:      p.Entity,
		Action:      p.Action,
		Description: p.Description,
		Key:         p.GetKey(),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (p *Permission) GetKey() string {
	return fmt.Sprintf("%s:%s:%s", p.Service, p.Entity, p.Action)
}
