package repository

import (
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

type SystemUserRepository interface {
	Create(systemUser *models.SystemUser) error
	GetByID(id uuid.UUID) (*models.SystemUser, error)
	GetByUserID(userID string) (*models.SystemUser, error)
	GetByEmail(email string) (*models.SystemUser, error)
	GetByName(name string) (*models.SystemUser, error)
	GetByApplicationName(applicationName string) ([]*models.SystemUser, error)
	List(activeOnly bool) ([]*models.SystemUser, error)
	Update(systemUser *models.SystemUser) error
	UpdateLastUsedAt(id uuid.UUID) error
	DeactivateExpired() (int64, error)
	Delete(id uuid.UUID) error
}

type systemUserRepository struct {
	db *gorm.DB
}

func NewSystemUserRepository(db *gorm.DB) SystemUserRepository {
	return &systemUserRepository{db: db}
}

func (r *systemUserRepository) Create(systemUser *models.SystemUser) error {
	return r.db.Create(systemUser).Error
}

func (r *systemUserRepository) GetByID(id uuid.UUID) (*models.SystemUser, error) {
	var systemUser models.SystemUser
	err := r.db.Where("id = ?", id).First(&systemUser).Error
	if err != nil {
		return nil, err
	}
	return &systemUser, nil
}

func (r *systemUserRepository) GetByUserID(userID string) (*models.SystemUser, error) {
	var systemUser models.SystemUser
	err := r.db.Where("user_id = ?", userID).First(&systemUser).Error
	if err != nil {
		return nil, err
	}
	return &systemUser, nil
}

func (r *systemUserRepository) GetByEmail(email string) (*models.SystemUser, error) {
	var systemUser models.SystemUser
	err := r.db.Where("email = ?", email).First(&systemUser).Error
	if err != nil {
		return nil, err
	}
	return &systemUser, nil
}

func (r *systemUserRepository) GetByName(name string) (*models.SystemUser, error) {
	var systemUser models.SystemUser
	err := r.db.Where("name = ?", name).First(&systemUser).Error
	if err != nil {
		return nil, err
	}
	return &systemUser, nil
}

func (r *systemUserRepository) GetByApplicationName(applicationName string) ([]*models.SystemUser, error) {
	var systemUsers []*models.SystemUser
	err := r.db.Where("application_name = ?", applicationName).
		Order("is_primary DESC, created_at DESC").
		Find(&systemUsers).Error
	if err != nil {
		return nil, err
	}
	return systemUsers, nil
}

func (r *systemUserRepository) List(activeOnly bool) ([]*models.SystemUser, error) {
	var systemUsers []*models.SystemUser
	query := r.db.Order("created_at DESC")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&systemUsers).Error
	if err != nil {
		return nil, err
	}
	return systemUsers, nil
}

func (r *systemUserRepository) Update(systemUser *models.SystemUser) error {
	return r.db.Save(systemUser).Error
}

func (r *systemUserRepository) UpdateLastUsedAt(id uuid.UUID) error {
	return r.db.Model(&models.SystemUser{}).
		Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error
}

func (r *systemUserRepository) DeactivateExpired() (int64, error) {
	result := r.db.Model(&models.SystemUser{}).
		Where("expires_at IS NOT NULL").
		Where("expires_at < NOW()").
		Where("is_active = ?", true).
		Update("is_active", false)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *systemUserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.SystemUser{}, id).Error
}
