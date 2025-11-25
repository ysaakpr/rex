package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

type InvitationRepository interface {
	Create(invitation *models.UserInvitation) error
	GetByID(id uuid.UUID) (*models.UserInvitation, error)
	GetByToken(token string) (*models.UserInvitation, error)
	GetByTenantID(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.UserInvitation, int64, error)
	GetByEmail(email string) ([]*models.UserInvitation, error)
	GetPendingByEmail(email string) ([]*models.UserInvitation, error)
	Update(invitation *models.UserInvitation) error
	UpdateStatus(id uuid.UUID, status models.InvitationStatus) error
	Delete(id uuid.UUID) error
	ExpireOldInvitations() error
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(invitation *models.UserInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *invitationRepository) GetByID(id uuid.UUID) (*models.UserInvitation, error) {
	var invitation models.UserInvitation
	err := r.db.
		Preload("Tenant").
		Preload("Role").
		Where("id = ?", id).
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) GetByToken(token string) (*models.UserInvitation, error) {
	var invitation models.UserInvitation
	err := r.db.
		Preload("Tenant").
		Preload("Role").
		Where("token = ?", token).
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) GetByTenantID(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.UserInvitation, int64, error) {
	var invitations []*models.UserInvitation
	var total int64

	query := r.db.Model(&models.UserInvitation{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pagination.Normalize()
	err := query.
		Preload("Tenant").
		Preload("Role").
		Offset(pagination.GetOffset()).
		Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&invitations).Error

	return invitations, total, err
}

func (r *invitationRepository) GetByEmail(email string) ([]*models.UserInvitation, error) {
	var invitations []*models.UserInvitation
	err := r.db.
		Preload("Tenant").
		Preload("Role").
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *invitationRepository) GetPendingByEmail(email string) ([]*models.UserInvitation, error) {
	var invitations []*models.UserInvitation
	err := r.db.
		Preload("Tenant").
		Preload("Role").
		Where("email = ? AND status = ? AND expires_at > ?", email, models.InvitationStatusPending, time.Now()).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *invitationRepository) Update(invitation *models.UserInvitation) error {
	return r.db.Save(invitation).Error
}

func (r *invitationRepository) UpdateStatus(id uuid.UUID, status models.InvitationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.InvitationStatusAccepted {
		now := time.Now()
		updates["accepted_at"] = &now
	}
	return r.db.Model(&models.UserInvitation{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *invitationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.UserInvitation{}, id).Error
}

func (r *invitationRepository) ExpireOldInvitations() error {
	return r.db.Model(&models.UserInvitation{}).
		Where("status = ? AND expires_at < ?", models.InvitationStatusPending, time.Now()).
		Update("status", models.InvitationStatusExpired).Error
}
