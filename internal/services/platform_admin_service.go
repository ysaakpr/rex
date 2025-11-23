package services

import (
	"errors"

	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/repository"
)

type PlatformAdminService interface {
	CreateAdmin(userID, createdBy string) (*models.PlatformAdmin, error)
	GetAdmin(userID string) (*models.PlatformAdmin, error)
	ListAdmins() ([]*models.PlatformAdmin, error)
	DeleteAdmin(userID string) error
	IsPlatformAdmin(userID string) (bool, error)
}

type platformAdminService struct {
	adminRepo repository.PlatformAdminRepository
}

func NewPlatformAdminService(adminRepo repository.PlatformAdminRepository) PlatformAdminService {
	return &platformAdminService{
		adminRepo: adminRepo,
	}
}

func (s *platformAdminService) CreateAdmin(userID, createdBy string) (*models.PlatformAdmin, error) {
	// Check if already exists
	existing, err := s.adminRepo.GetByUserID(userID)
	if err == nil && existing != nil {
		return nil, errors.New("user is already a platform admin")
	}

	admin := &models.PlatformAdmin{
		UserID:    userID,
		CreatedBy: createdBy,
	}

	if err := s.adminRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *platformAdminService) GetAdmin(userID string) (*models.PlatformAdmin, error) {
	return s.adminRepo.GetByUserID(userID)
}

func (s *platformAdminService) ListAdmins() ([]*models.PlatformAdmin, error) {
	return s.adminRepo.List()
}

func (s *platformAdminService) DeleteAdmin(userID string) error {
	return s.adminRepo.Delete(userID)
}

func (s *platformAdminService) IsPlatformAdmin(userID string) (bool, error) {
	return s.adminRepo.IsPlatformAdmin(userID)
}
