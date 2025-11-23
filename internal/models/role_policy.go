package models

import (
	"time"

	"github.com/google/uuid"
)

// RolePolicy represents the junction between roles and policies (was RelationRole)
type RolePolicy struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	PolicyID  uuid.UUID `gorm:"type:uuid;not null" json:"policy_id"`
	CreatedAt time.Time `json:"created_at"`

	// Associations
	Role   Role   `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Policy Policy `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
}

func (RolePolicy) TableName() string {
	return "role_policies"
}
