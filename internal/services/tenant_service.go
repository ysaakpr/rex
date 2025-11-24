package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/jobs"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/repository"
	"gorm.io/gorm"
)

type TenantService interface {
	CreateTenant(input *models.CreateTenantInput, creatorID string) (*models.Tenant, error)
	CreateManagedTenant(input *models.CreateTenantInput, adminEmail string, creatorID string) (*models.Tenant, error)
	GetTenant(id uuid.UUID) (*models.Tenant, error)
	GetTenantBySlug(slug string) (*models.Tenant, error)
	GetUserTenants(userID string, pagination *models.PaginationParams) ([]*models.Tenant, int64, error)
	GetAllTenants(pagination *models.PaginationParams) ([]*models.Tenant, int64, error)
	UpdateTenant(id uuid.UUID, input *models.UpdateTenantInput) (*models.Tenant, error)
	DeleteTenant(id uuid.UUID) error
	GetTenantStatus(id uuid.UUID) (models.TenantStatus, error)
}

type tenantService struct {
	tenantRepo     repository.TenantRepository
	memberRepo     repository.MemberRepository
	invitationRepo repository.InvitationRepository
	rbacRepo       repository.RBACRepository
	jobClient      jobs.Client
}

func NewTenantService(
	tenantRepo repository.TenantRepository,
	memberRepo repository.MemberRepository,
	invitationRepo repository.InvitationRepository,
	rbacRepo repository.RBACRepository,
	jobClient jobs.Client,
) TenantService {
	return &tenantService{
		tenantRepo:     tenantRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
		rbacRepo:       rbacRepo,
		jobClient:      jobClient,
	}
}

func (s *tenantService) CreateTenant(input *models.CreateTenantInput, creatorID string) (*models.Tenant, error) {
	// Check if slug already exists
	existing, err := s.tenantRepo.GetBySlug(input.Slug)
	if err == nil && existing != nil {
		return nil, errors.New("tenant slug already exists")
	}

	// Create tenant
	tenant := &models.Tenant{
		Name:      input.Name,
		Slug:      normalizeSlug(input.Slug),
		Status:    models.TenantStatusPending,
		Metadata:  input.Metadata,
		CreatedBy: creatorID,
	}

	if err := s.tenantRepo.Create(tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Get Admin role
	adminRole, err := s.rbacRepo.GetRoleByName("Admin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin role: %w", err)
	}

	// Add creator as admin member
	member := &models.TenantMember{
		TenantID: tenant.ID,
		UserID:   creatorID,
		RoleID:   adminRole.ID,
		Status:   models.MemberStatusActive,
		JoinedAt: time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to add creator as admin: %w", err)
	}

	// Enqueue tenant initialization job
	if err := s.jobClient.EnqueueTenantInitialization(tenant.ID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to enqueue tenant initialization: %v\n", err)
	}

	return tenant, nil
}

func (s *tenantService) CreateManagedTenant(input *models.CreateTenantInput, adminEmail string, creatorID string) (*models.Tenant, error) {
	// Check if slug already exists
	existing, err := s.tenantRepo.GetBySlug(input.Slug)
	if err == nil && existing != nil {
		return nil, errors.New("tenant slug already exists")
	}

	// Create tenant
	tenant := &models.Tenant{
		Name:      input.Name,
		Slug:      normalizeSlug(input.Slug),
		Status:    models.TenantStatusPending,
		Metadata:  input.Metadata,
		CreatedBy: creatorID,
	}

	if err := s.tenantRepo.Create(tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Get Admin role
	adminRole, err := s.rbacRepo.GetRoleByName("Admin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin role: %w", err)
	}

	// Create invitation for the admin user
	invitationToken := generateInvitationToken()
	invitation := &models.UserInvitation{
		TenantID:  tenant.ID,
		Email:     adminEmail,
		InvitedBy: creatorID,
		RoleID:    adminRole.ID,
		Token:     invitationToken,
		Status:    models.InvitationStatusPending,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Enqueue invitation email job
	if err := s.jobClient.EnqueueUserInvitation(invitation.ID); err != nil {
		fmt.Printf("failed to enqueue invitation email: %v\n", err)
	}

	// Note: Tenant initialization will be triggered when admin accepts invitation

	return tenant, nil
}

func (s *tenantService) GetTenant(id uuid.UUID) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}
	return tenant, nil
}

func (s *tenantService) GetTenantBySlug(slug string) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}
	return tenant, nil
}

func (s *tenantService) GetUserTenants(userID string, pagination *models.PaginationParams) ([]*models.Tenant, int64, error) {
	// Get all memberships for the user
	members, err := s.memberRepo.GetByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	if len(members) == 0 {
		return []*models.Tenant{}, 0, nil
	}

	// Extract tenant IDs
	tenantIDs := make([]uuid.UUID, len(members))
	for i, member := range members {
		tenantIDs[i] = member.TenantID
	}

	// For simplicity, we'll return the tenants the user created
	// In a full implementation, you'd query by tenant IDs
	return s.tenantRepo.GetByCreatorID(userID, pagination)
}

// GetAllTenants returns all tenants in the system (for platform admins)
func (s *tenantService) GetAllTenants(pagination *models.PaginationParams) ([]*models.Tenant, int64, error) {
	return s.tenantRepo.List(pagination)
}

func (s *tenantService) UpdateTenant(id uuid.UUID, input *models.UpdateTenantInput) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	// Update fields
	if input.Name != nil {
		tenant.Name = *input.Name
	}
	if input.Status != nil {
		tenant.Status = *input.Status
	}
	if input.Metadata != nil {
		tenant.Metadata = input.Metadata
	}

	if err := s.tenantRepo.Update(tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) DeleteTenant(id uuid.UUID) error {
	tenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("tenant not found")
		}
		return err
	}

	// Soft delete
	tenant.Status = models.TenantStatusDeleted
	return s.tenantRepo.Update(tenant)
}

func (s *tenantService) GetTenantStatus(id uuid.UUID) (models.TenantStatus, error) {
	tenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("tenant not found")
		}
		return "", err
	}
	return tenant.Status, nil
}

func normalizeSlug(slug string) string {
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func generateInvitationToken() string {
	return uuid.New().String()
}
