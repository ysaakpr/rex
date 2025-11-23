package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/config"
	"github.com/ysaakpr/rex/internal/jobs"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/repository"
	"gorm.io/gorm"
)

type InvitationService interface {
	CreateInvitation(tenantID uuid.UUID, input *models.CreateInvitationInput, invitedBy string) (*models.UserInvitation, error)
	GetInvitation(id uuid.UUID) (*models.UserInvitation, error)
	GetTenantInvitations(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.UserInvitation, int64, error)
	AcceptInvitation(token string, userID string) (*models.TenantMember, error)
	CancelInvitation(id uuid.UUID) error
	CheckAndAcceptPendingInvitations(email string, userID string) ([]*models.TenantMember, error)
}

type invitationService struct {
	invitationRepo repository.InvitationRepository
	memberRepo     repository.MemberRepository
	tenantRepo     repository.TenantRepository
	rbacRepo       repository.RBACRepository
	jobClient      jobs.Client
	cfg            *config.Config
}

func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	memberRepo repository.MemberRepository,
	tenantRepo repository.TenantRepository,
	rbacRepo repository.RBACRepository,
	jobClient jobs.Client,
	cfg *config.Config,
) InvitationService {
	return &invitationService{
		invitationRepo: invitationRepo,
		memberRepo:     memberRepo,
		tenantRepo:     tenantRepo,
		rbacRepo:       rbacRepo,
		jobClient:      jobClient,
		cfg:            cfg,
	}
}

func (s *invitationService) CreateInvitation(tenantID uuid.UUID, input *models.CreateInvitationInput, invitedBy string) (*models.UserInvitation, error) {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	// Validate role exists
	_, err = s.rbacRepo.GetRoleByID(input.RoleID)
	if err != nil {
		return nil, errors.New("invalid role")
	}

	// Check if there's already a pending invitation for this email
	pendingInvitations, err := s.invitationRepo.GetPendingByEmail(input.Email)
	if err == nil && len(pendingInvitations) > 0 {
		for _, inv := range pendingInvitations {
			if inv.TenantID == tenantID {
				return nil, errors.New("user already has a pending invitation for this tenant")
			}
		}
	}

	// Generate invitation token
	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create invitation
	invitation := &models.UserInvitation{
		TenantID:  tenant.ID,
		Email:     input.Email,
		InvitedBy: invitedBy,
		RoleID:    input.RoleID,
		Token:     token,
		Status:    models.InvitationStatusPending,
		ExpiresAt: time.Now().Add(s.cfg.GetInvitationExpiry()),
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Enqueue invitation email job
	if err := s.jobClient.EnqueueUserInvitation(invitation.ID); err != nil {
		fmt.Printf("failed to enqueue invitation email: %v\n", err)
	}

	return s.invitationRepo.GetByID(invitation.ID)
}

func (s *invitationService) GetInvitation(id uuid.UUID) (*models.UserInvitation, error) {
	invitation, err := s.invitationRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return invitation, nil
}

func (s *invitationService) GetTenantInvitations(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.UserInvitation, int64, error) {
	return s.invitationRepo.GetByTenantID(tenantID, pagination)
}

func (s *invitationService) AcceptInvitation(token string, userID string) (*models.TenantMember, error) {
	// Get invitation by token
	invitation, err := s.invitationRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}

	// Check if invitation can be accepted
	if !invitation.CanAccept() {
		return nil, errors.New("invitation cannot be accepted (expired or already used)")
	}

	// Check if user is already a member
	existing, err := s.memberRepo.GetByTenantAndUser(invitation.TenantID, userID)
	if err == nil && existing != nil {
		return nil, errors.New("user is already a member of this tenant")
	}

	// Create member
	member := &models.TenantMember{
		TenantID:  invitation.TenantID,
		UserID:    userID,
		RoleID:    invitation.RoleID,
		Status:    models.MemberStatusActive,
		InvitedBy: &invitation.InvitedBy,
		JoinedAt:  time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = models.InvitationStatusAccepted
	invitation.AcceptedAt = &now
	if err := s.invitationRepo.Update(invitation); err != nil {
		fmt.Printf("failed to update invitation status: %v\n", err)
	}

	// If this was a managed tenant creation (first admin), trigger tenant initialization
	tenant, _ := s.tenantRepo.GetByID(invitation.TenantID)
	if tenant != nil && tenant.Status == models.TenantStatusPending {
		if err := s.jobClient.EnqueueTenantInitialization(tenant.ID); err != nil {
			fmt.Printf("failed to enqueue tenant initialization: %v\n", err)
		}
	}

	return s.memberRepo.GetByID(member.ID)
}

func (s *invitationService) CancelInvitation(id uuid.UUID) error {
	invitation, err := s.invitationRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invitation not found")
		}
		return err
	}

	if invitation.Status != models.InvitationStatusPending {
		return errors.New("only pending invitations can be cancelled")
	}

	return s.invitationRepo.UpdateStatus(id, models.InvitationStatusCancelled)
}

func (s *invitationService) CheckAndAcceptPendingInvitations(email string, userID string) ([]*models.TenantMember, error) {
	// Get all pending invitations for this email
	invitations, err := s.invitationRepo.GetPendingByEmail(email)
	if err != nil {
		return nil, err
	}

	var members []*models.TenantMember
	for _, invitation := range invitations {
		if invitation.CanAccept() {
			member, err := s.AcceptInvitation(invitation.Token, userID)
			if err != nil {
				fmt.Printf("failed to auto-accept invitation %s: %v\n", invitation.ID, err)
				continue
			}
			members = append(members, member)
		}
	}

	return members, nil
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
