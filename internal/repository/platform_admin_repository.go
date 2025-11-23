package repository

import (
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

type PlatformAdminRepository interface {
	Create(admin *models.PlatformAdmin) error
	GetByUserID(userID string) (*models.PlatformAdmin, error)
	List() ([]*models.PlatformAdmin, error)
	Delete(userID string) error
	IsPlatformAdmin(userID string) (bool, error)
}

type platformAdminRepository struct {
	db *gorm.DB
}

func NewPlatformAdminRepository(db *gorm.DB) PlatformAdminRepository {
	return &platformAdminRepository{db: db}
}

func (r *platformAdminRepository) Create(admin *models.PlatformAdmin) error {
	return r.db.Create(admin).Error
}

func (r *platformAdminRepository) GetByUserID(userID string) (*models.PlatformAdmin, error) {
	var admin models.PlatformAdmin
	if err := r.db.Where("user_id = ?", userID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *platformAdminRepository) List() ([]*models.PlatformAdmin, error) {
	var admins []*models.PlatformAdmin
	if err := r.db.Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *platformAdminRepository) Delete(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.PlatformAdmin{}).Error
}

func (r *platformAdminRepository) IsPlatformAdmin(userID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PlatformAdmin{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
