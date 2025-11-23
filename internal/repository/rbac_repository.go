package repository

import (
	"github.com/google/uuid"
	"github.com/vyshakhp/utm-backend/internal/models"
	"gorm.io/gorm"
)

type RBACRepository interface {
	// Roles (was Relations - user's role in tenant: Admin, Writer, etc.)
	CreateRole(role *models.Role) error
	GetRoleByID(id uuid.UUID) (*models.Role, error)
	GetRoleByName(name string, tenantID *uuid.UUID) (*models.Role, error)
	GetRoleWithPolicies(id uuid.UUID) (*models.Role, error)
	ListRoles(tenantID *uuid.UUID) ([]*models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(id uuid.UUID) error

	// Policies (was Roles - group of permissions)
	CreatePolicy(policy *models.Policy) error
	GetPolicyByID(id uuid.UUID) (*models.Policy, error)
	GetPolicyWithPermissions(id uuid.UUID) (*models.Policy, error)
	ListPolicies(tenantID *uuid.UUID) ([]*models.Policy, error)
	UpdatePolicy(policy *models.Policy) error
	DeletePolicy(id uuid.UUID) error

	// Permissions
	CreatePermission(permission *models.Permission) error
	GetPermissionByID(id uuid.UUID) (*models.Permission, error)
	GetPermissionByKey(service, entity, action string) (*models.Permission, error)
	ListPermissions() ([]*models.Permission, error)
	ListPermissionsByService(service string) ([]*models.Permission, error)
	DeletePermission(id uuid.UUID) error

	// Policy-Permission assignments
	AssignPermissionsToPolicy(policyID uuid.UUID, permissionIDs []uuid.UUID) error
	RevokePermissionFromPolicy(policyID uuid.UUID, permissionID uuid.UUID) error
	GetPolicyPermissions(policyID uuid.UUID) ([]*models.Permission, error)

	// Authorization queries
	GetUserPermissions(tenantID uuid.UUID, userID string) ([]*models.Permission, error)
	CheckUserPermission(tenantID uuid.UUID, userID string, service, entity, action string) (bool, error)

	// Role-Policy assignments (was Relation-Role)
	AssignPoliciesToRole(roleID uuid.UUID, policyIDs []uuid.UUID) error
	RevokePolicyFromRole(roleID uuid.UUID, policyID uuid.UUID) error
	GetRolePolicies(roleID uuid.UUID) ([]*models.Policy, error)
}

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{db: db}
}

// Roles (was Relations)
func (r *rbacRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *rbacRepository) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *rbacRepository) GetRoleByName(name string, tenantID *uuid.UUID) (*models.Role, error) {
	var role models.Role
	query := r.db.Where("name = ?", name)
	if tenantID != nil {
		query = query.Where("(tenant_id = ? OR tenant_id IS NULL)", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.First(&role).Error
	return &role, err
}

func (r *rbacRepository) GetRoleWithPolicies(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Policies").Preload("Policies.Permissions").
		Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *rbacRepository) ListRoles(tenantID *uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	query := r.db.Model(&models.Role{}).Preload("Policies")
	if tenantID != nil {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	} else {
		query = query.Where("is_system = ?", true)
	}
	err := query.Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) UpdateRole(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *rbacRepository) DeleteRole(id uuid.UUID) error {
	return r.db.Delete(&models.Role{}, id).Error
}

// Policies (was Roles)
func (r *rbacRepository) CreatePolicy(policy *models.Policy) error {
	return r.db.Create(policy).Error
}

func (r *rbacRepository) GetPolicyByID(id uuid.UUID) (*models.Policy, error) {
	var policy models.Policy
	err := r.db.Where("id = ?", id).First(&policy).Error
	return &policy, err
}

func (r *rbacRepository) GetPolicyWithPermissions(id uuid.UUID) (*models.Policy, error) {
	var policy models.Policy
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&policy).Error
	return &policy, err
}

func (r *rbacRepository) ListPolicies(tenantID *uuid.UUID) ([]*models.Policy, error) {
	var policies []*models.Policy
	query := r.db.Model(&models.Policy{}).Preload("Permissions")
	if tenantID != nil {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	} else {
		query = query.Where("is_system = ?", true)
	}
	err := query.Order("name ASC").Find(&policies).Error
	return policies, err
}

func (r *rbacRepository) UpdatePolicy(policy *models.Policy) error {
	return r.db.Save(policy).Error
}

func (r *rbacRepository) DeletePolicy(id uuid.UUID) error {
	return r.db.Delete(&models.Policy{}, id).Error
}

// Permissions
func (r *rbacRepository) CreatePermission(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *rbacRepository) GetPermissionByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	return &permission, err
}

func (r *rbacRepository) GetPermissionByKey(service, entity, action string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("service = ? AND entity = ? AND action = ?", service, entity, action).
		First(&permission).Error
	return &permission, err
}

func (r *rbacRepository) ListPermissions() ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Order("service ASC, entity ASC, action ASC").Find(&permissions).Error
	return permissions, err
}

func (r *rbacRepository) ListPermissionsByService(service string) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Where("service = ?", service).
		Order("entity ASC, action ASC").
		Find(&permissions).Error
	return permissions, err
}

func (r *rbacRepository) DeletePermission(id uuid.UUID) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

// Policy-Permission assignments
func (r *rbacRepository) AssignPermissionsToPolicy(policyID uuid.UUID, permissionIDs []uuid.UUID) error {
	policy := &models.Policy{ID: policyID}
	return r.db.Model(policy).Association("Permissions").Append(convertToPermissions(permissionIDs))
}

func (r *rbacRepository) RevokePermissionFromPolicy(policyID uuid.UUID, permissionID uuid.UUID) error {
	policy := &models.Policy{ID: policyID}
	permission := &models.Permission{ID: permissionID}
	return r.db.Model(policy).Association("Permissions").Delete(permission)
}

func (r *rbacRepository) GetPolicyPermissions(policyID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.
		Joins("JOIN policy_permissions ON policy_permissions.permission_id = permissions.id").
		Where("policy_permissions.policy_id = ?", policyID).
		Find(&permissions).Error
	return permissions, err
}

// Authorization queries
func (r *rbacRepository) GetUserPermissions(tenantID uuid.UUID, userID string) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Raw(`
		SELECT DISTINCT p.*
		FROM permissions p
		INNER JOIN policy_permissions pp ON pp.permission_id = p.id
		INNER JOIN role_policies rp ON rp.policy_id = pp.policy_id
		INNER JOIN tenant_members tm ON tm.role_id = rp.role_id
		WHERE tm.tenant_id = ? AND tm.user_id = ? AND tm.status = 'active'
	`, tenantID, userID).Scan(&permissions).Error
	return permissions, err
}

func (r *rbacRepository) CheckUserPermission(tenantID uuid.UUID, userID string, service, entity, action string) (bool, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(DISTINCT p.id)
		FROM permissions p
		INNER JOIN policy_permissions pp ON pp.permission_id = p.id
		INNER JOIN role_policies rp ON rp.policy_id = pp.policy_id
		INNER JOIN tenant_members tm ON tm.role_id = rp.role_id
		WHERE tm.tenant_id = ? 
		  AND tm.user_id = ? 
		  AND tm.status = 'active'
		  AND p.service = ?
		  AND p.entity = ?
		  AND p.action = ?
	`, tenantID, userID, service, entity, action).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func convertToPermissions(permissionIDs []uuid.UUID) []*models.Permission {
	permissions := make([]*models.Permission, len(permissionIDs))
	for i, id := range permissionIDs {
		permissions[i] = &models.Permission{ID: id}
	}
	return permissions
}

// AssignPoliciesToRole assigns multiple policies to a role
func (r *rbacRepository) AssignPoliciesToRole(roleID uuid.UUID, policyIDs []uuid.UUID) error {
	// Add new policy assignments (don't delete existing ones)
	for _, policyID := range policyIDs {
		// Check if the assignment already exists
		var existing models.RolePolicy
		err := r.db.Where("role_id = ? AND policy_id = ?", roleID, policyID).
			First(&existing).Error

		if err == nil {
			// Assignment already exists, skip
			continue
		}

		// Create new assignment
		rolePolicy := &models.RolePolicy{
			RoleID:   roleID,
			PolicyID: policyID,
		}
		if err := r.db.Create(rolePolicy).Error; err != nil {
			return err
		}
	}

	return nil
}

// RevokePolicyFromRole removes a policy from a role
func (r *rbacRepository) RevokePolicyFromRole(roleID uuid.UUID, policyID uuid.UUID) error {
	return r.db.Where("role_id = ? AND policy_id = ?", roleID, policyID).
		Delete(&models.RolePolicy{}).Error
}

// GetRolePolicies retrieves all policies associated with a role
func (r *rbacRepository) GetRolePolicies(roleID uuid.UUID) ([]*models.Policy, error) {
	var policies []*models.Policy
	err := r.db.Table("policies").
		Select("policies.*").
		Joins("JOIN role_policies ON policies.id = role_policies.policy_id").
		Where("role_policies.role_id = ?", roleID).
		Preload("Permissions").
		Find(&policies).Error

	if err != nil {
		return nil, err
	}

	return policies, nil
}
