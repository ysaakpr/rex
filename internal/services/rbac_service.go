package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/repository"
	"gorm.io/gorm"
)

type RBACService interface {
	// Roles (was Relations - user's role in tenant: Admin, Writer, etc.)
	CreateRole(input *models.CreateRoleInput) (*models.Role, error)
	GetRole(id uuid.UUID) (*models.Role, error)
	ListRoles(tenantID *uuid.UUID) ([]*models.Role, error)
	UpdateRole(id uuid.UUID, input *models.UpdateRoleInput) (*models.Role, error)
	DeleteRole(id uuid.UUID) error

	// Policies (was Roles - group of permissions)
	CreatePolicy(input *models.CreatePolicyInput) (*models.Policy, error)
	GetPolicy(id uuid.UUID) (*models.Policy, error)
	ListPolicies(tenantID *uuid.UUID) ([]*models.Policy, error)
	UpdatePolicy(id uuid.UUID, input *models.UpdatePolicyInput) (*models.Policy, error)
	DeletePolicy(id uuid.UUID) error

	// Permissions
	CreatePermission(input *models.CreatePermissionInput) (*models.Permission, error)
	GetPermission(id uuid.UUID) (*models.Permission, error)
	ListPermissions() ([]*models.Permission, error)
	ListPermissionsByService(service string) ([]*models.Permission, error)
	DeletePermission(id uuid.UUID) error

	// Policy-Permission assignments
	AssignPermissionsToPolicy(policyID uuid.UUID, permissionIDs []uuid.UUID) error
	RevokePermissionFromPolicy(policyID uuid.UUID, permissionID uuid.UUID) error

	// Authorization
	CheckUserPermission(tenantID uuid.UUID, userID string, service, entity, action string) (bool, error)
	GetUserPermissions(tenantID uuid.UUID, userID string) ([]*models.Permission, error)

	// Role-Policy assignments (was Relation-Role)
	AssignPoliciesToRole(roleID uuid.UUID, policyIDs []uuid.UUID) error
	RevokePolicyFromRole(roleID uuid.UUID, policyID uuid.UUID) error
	GetRolePolicies(roleID uuid.UUID) ([]*models.Policy, error)
}

type rbacService struct {
	rbacRepo repository.RBACRepository
}

func NewRBACService(rbacRepo repository.RBACRepository) RBACService {
	return &rbacService{
		rbacRepo: rbacRepo,
	}
}

// Roles (was Relations)
func (s *rbacService) CreateRole(input *models.CreateRoleInput) (*models.Role, error) {
	// Check if role already exists
	existing, err := s.rbacRepo.GetRoleByName(input.Name, input.TenantID)
	if err == nil && existing != nil {
		return nil, errors.New("role with this name already exists")
	}

	role := &models.Role{
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
		TenantID:    input.TenantID,
		IsSystem:    input.TenantID == nil,
	}

	if err := s.rbacRepo.CreateRole(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *rbacService) GetRole(id uuid.UUID) (*models.Role, error) {
	role, err := s.rbacRepo.GetRoleWithPolicies(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (s *rbacService) ListRoles(tenantID *uuid.UUID) ([]*models.Role, error) {
	roles, err := s.rbacRepo.ListRoles(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

func (s *rbacService) UpdateRole(id uuid.UUID, input *models.UpdateRoleInput) (*models.Role, error) {
	role, err := s.rbacRepo.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	if input.Name != nil {
		role.Name = *input.Name
	}
	if input.Description != nil {
		role.Description = *input.Description
	}

	if err := s.rbacRepo.UpdateRole(role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

func (s *rbacService) DeleteRole(id uuid.UUID) error {
	_, err := s.rbacRepo.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if err := s.rbacRepo.DeleteRole(id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// Policies (was Roles)
func (s *rbacService) CreatePolicy(input *models.CreatePolicyInput) (*models.Policy, error) {
	policy := &models.Policy{
		Name:        input.Name,
		Description: input.Description,
		TenantID:    input.TenantID,
		IsSystem:    input.TenantID == nil,
	}

	if err := s.rbacRepo.CreatePolicy(policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return policy, nil
}

func (s *rbacService) GetPolicy(id uuid.UUID) (*models.Policy, error) {
	policy, err := s.rbacRepo.GetPolicyWithPermissions(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return policy, nil
}

func (s *rbacService) ListPolicies(tenantID *uuid.UUID) ([]*models.Policy, error) {
	policies, err := s.rbacRepo.ListPolicies(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}

func (s *rbacService) UpdatePolicy(id uuid.UUID, input *models.UpdatePolicyInput) (*models.Policy, error) {
	policy, err := s.rbacRepo.GetPolicyByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if input.Name != nil {
		policy.Name = *input.Name
	}
	if input.Description != nil {
		policy.Description = *input.Description
	}

	if err := s.rbacRepo.UpdatePolicy(policy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return policy, nil
}

func (s *rbacService) DeletePolicy(id uuid.UUID) error {
	_, err := s.rbacRepo.GetPolicyByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy not found")
		}
		return fmt.Errorf("failed to get policy: %w", err)
	}

	if err := s.rbacRepo.DeletePolicy(id); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// Permissions
func (s *rbacService) CreatePermission(input *models.CreatePermissionInput) (*models.Permission, error) {
	// Check if permission already exists
	existing, err := s.rbacRepo.GetPermissionByKey(input.Service, input.Entity, input.Action)
	if err == nil && existing != nil {
		return nil, errors.New("permission with this service, entity, and action already exists")
	}

	permission := &models.Permission{
		Service:     input.Service,
		Entity:      input.Entity,
		Action:      input.Action,
		Description: input.Description,
	}

	if err := s.rbacRepo.CreatePermission(permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *rbacService) GetPermission(id uuid.UUID) (*models.Permission, error) {
	permission, err := s.rbacRepo.GetPermissionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (s *rbacService) ListPermissions() ([]*models.Permission, error) {
	permissions, err := s.rbacRepo.ListPermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

func (s *rbacService) ListPermissionsByService(service string) ([]*models.Permission, error) {
	permissions, err := s.rbacRepo.ListPermissionsByService(service)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by service: %w", err)
	}
	return permissions, nil
}

func (s *rbacService) DeletePermission(id uuid.UUID) error {
	_, err := s.rbacRepo.GetPermissionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	if err := s.rbacRepo.DeletePermission(id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// Policy-Permission assignments
func (s *rbacService) AssignPermissionsToPolicy(policyID uuid.UUID, permissionIDs []uuid.UUID) error {
	// Verify policy exists
	_, err := s.rbacRepo.GetPolicyByID(policyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy not found")
		}
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Verify all permissions exist
	for _, permID := range permissionIDs {
		_, err := s.rbacRepo.GetPermissionByID(permID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("permission %s not found", permID)
			}
			return fmt.Errorf("failed to verify permission: %w", err)
		}
	}

	if err := s.rbacRepo.AssignPermissionsToPolicy(policyID, permissionIDs); err != nil {
		return fmt.Errorf("failed to assign permissions to policy: %w", err)
	}

	return nil
}

func (s *rbacService) RevokePermissionFromPolicy(policyID uuid.UUID, permissionID uuid.UUID) error {
	if err := s.rbacRepo.RevokePermissionFromPolicy(policyID, permissionID); err != nil {
		return fmt.Errorf("failed to revoke permission from policy: %w", err)
	}
	return nil
}

// Authorization
func (s *rbacService) CheckUserPermission(tenantID uuid.UUID, userID string, service, entity, action string) (bool, error) {
	hasPermission, err := s.rbacRepo.CheckUserPermission(tenantID, userID, service, entity, action)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}
	return hasPermission, nil
}

func (s *rbacService) GetUserPermissions(tenantID uuid.UUID, userID string) ([]*models.Permission, error) {
	permissions, err := s.rbacRepo.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// Role-Policy assignments (was Relation-Role)
func (s *rbacService) AssignPoliciesToRole(roleID uuid.UUID, policyIDs []uuid.UUID) error {
	// Verify role exists
	_, err := s.rbacRepo.GetRoleByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Verify all policies exist
	for _, policyID := range policyIDs {
		_, err := s.rbacRepo.GetPolicyByID(policyID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("policy %s not found", policyID)
			}
			return fmt.Errorf("failed to verify policy: %w", err)
		}
	}

	if err := s.rbacRepo.AssignPoliciesToRole(roleID, policyIDs); err != nil {
		return fmt.Errorf("failed to assign policies to role: %w", err)
	}

	return nil
}

func (s *rbacService) RevokePolicyFromRole(roleID uuid.UUID, policyID uuid.UUID) error {
	if err := s.rbacRepo.RevokePolicyFromRole(roleID, policyID); err != nil {
		return fmt.Errorf("failed to revoke policy from role: %w", err)
	}
	return nil
}

func (s *rbacService) GetRolePolicies(roleID uuid.UUID) ([]*models.Policy, error) {
	policies, err := s.rbacRepo.GetRolePolicies(roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role policies: %w", err)
	}
	return policies, nil
}
