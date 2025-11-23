package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vyshakhp/utm-backend/internal/models"
	"github.com/vyshakhp/utm-backend/internal/repository"
	"gorm.io/gorm"
)

type MemberService interface {
	AddMember(tenantID uuid.UUID, input *models.AddMemberInput, invitedBy string) (*models.TenantMember, error)
	GetMember(tenantID uuid.UUID, userID string) (*models.TenantMember, error)
	GetTenantMembers(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.TenantMember, int64, error)
	UpdateMember(memberID uuid.UUID, input *models.UpdateMemberInput) (*models.TenantMember, error)
	RemoveMember(memberID uuid.UUID) error
	AssignRolesToMember(memberID uuid.UUID, roleIDs []uuid.UUID) error
	RemoveRoleFromMember(memberID uuid.UUID, roleID uuid.UUID) error
	GetMemberWithPermissions(memberID uuid.UUID) (*models.TenantMember, error)
}

type memberService struct {
	memberRepo repository.MemberRepository
	tenantRepo repository.TenantRepository
	rbacRepo   repository.RBACRepository
}

func NewMemberService(
	memberRepo repository.MemberRepository,
	tenantRepo repository.TenantRepository,
	rbacRepo repository.RBACRepository,
) MemberService {
	return &memberService{
		memberRepo: memberRepo,
		tenantRepo: tenantRepo,
		rbacRepo:   rbacRepo,
	}
}

func (s *memberService) AddMember(tenantID uuid.UUID, input *models.AddMemberInput, invitedBy string) (*models.TenantMember, error) {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	// Check if user is already a member
	existing, err := s.memberRepo.GetByTenantAndUser(tenantID, input.UserID)
	if err == nil && existing != nil {
		return nil, errors.New("user is already a member of this tenant")
	}

	// Validate role exists
	role, err := s.rbacRepo.GetRoleByID(input.RoleID)
	if err != nil {
		return nil, errors.New("invalid role")
	}

	// Check if relation belongs to tenant or is system-wide
	if relation.TenantID != nil && *relation.TenantID != tenantID {
		return nil, errors.New("relation does not belong to this tenant")
	}

	// Create member
	member := &models.TenantMember{
		TenantID:  tenant.ID,
		UserID:    input.UserID,
		RoleID:    input.RoleID,
		Status:    models.MemberStatusActive,
		InvitedBy: &invitedBy,
		JoinedAt:   time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return s.memberRepo.GetByID(member.ID)
}

func (s *memberService) GetMember(tenantID uuid.UUID, userID string) (*models.TenantMember, error) {
	member, err := s.memberRepo.GetByTenantAndUser(tenantID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}
	return member, nil
}

func (s *memberService) GetTenantMembers(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.TenantMember, int64, error) {
	return s.memberRepo.GetByTenantID(tenantID, pagination)
}

func (s *memberService) UpdateMember(memberID uuid.UUID, input *models.UpdateMemberInput) (*models.TenantMember, error) {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}

	// Update fields
	if input.RoleID != nil {
		// Validate role exists
		_, err := s.rbacRepo.GetRoleByID(*input.RoleID)
		if err != nil {
			return nil, errors.New("invalid role")
		}
		member.RoleID = *input.RoleID
	}

	if input.Status != nil {
		member.Status = *input.Status
	}

	if err := s.memberRepo.Update(member); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	return s.memberRepo.GetByID(memberID)
}

func (s *memberService) RemoveMember(memberID uuid.UUID) error {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("member not found")
		}
		return err
	}

	// Check if this is the last admin (optional business logic)
	// For now, we'll allow removal

	return s.memberRepo.Delete(member.ID)
}

func (s *memberService) AssignRolesToMember(memberID uuid.UUID, roleIDs []uuid.UUID) error {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("member not found")
		}
		return err
	}

	// Validate all roles exist
	for _, roleID := range roleIDs {
		_, err := s.rbacRepo.GetRoleByID(roleID)
		if err != nil {
			return fmt.Errorf("invalid role: %s", roleID)
		}
	}

	return s.memberRepo.AssignRoles(member.ID, roleIDs)
}

func (s *memberService) RemoveRoleFromMember(memberID uuid.UUID, roleID uuid.UUID) error {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("member not found")
		}
		return err
	}

	return s.memberRepo.RemoveRole(member.ID, roleID)
}

func (s *memberService) GetMemberWithPermissions(memberID uuid.UUID) (*models.TenantMember, error) {
	return s.memberRepo.GetMemberWithRoles(memberID)
}
