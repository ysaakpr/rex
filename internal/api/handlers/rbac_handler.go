package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vyshakhp/utm-backend/internal/models"
	"github.com/vyshakhp/utm-backend/internal/pkg/response"
	"github.com/vyshakhp/utm-backend/internal/services"
)

type RBACHandler struct {
	rbacService services.RBACService
}

func NewRBACHandler(rbacService services.RBACService) *RBACHandler {
	return &RBACHandler{
		rbacService: rbacService,
	}
}

// ============================================================================
// Roles (was Relations - user's role in tenant: Admin, Writer, etc.)
// ============================================================================

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var input models.CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	role, err := h.rbacService.CreateRole(&input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, role.ToResponse())
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		id, err := uuid.Parse(tenantIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid tenant_id", err)
			return
		}
		tenantID = &id
	}

	roles, err := h.rbacService.ListRoles(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list roles", err)
		return
	}

	roleResponses := make([]*models.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = role.ToResponse()
	}

	response.Success(c, http.StatusOK, roleResponses)
}

func (h *RBACHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	role, err := h.rbacService.GetRole(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Role not found", err)
		return
	}

	response.Success(c, http.StatusOK, role.ToResponse())
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	var input models.UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	role, err := h.rbacService.UpdateRole(id, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, role.ToResponse())
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	if err := h.rbacService.DeleteRole(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// ============================================================================
// Policies (was Roles - group of permissions)
// ============================================================================

func (h *RBACHandler) CreatePolicy(c *gin.Context) {
	var input models.CreatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	policy, err := h.rbacService.CreatePolicy(&input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, policy.ToResponse())
}

func (h *RBACHandler) ListPolicies(c *gin.Context) {
	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		id, err := uuid.Parse(tenantIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid tenant_id", err)
			return
		}
		tenantID = &id
	}

	policies, err := h.rbacService.ListPolicies(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list policies", err)
		return
	}

	policyResponses := make([]*models.PolicyResponse, len(policies))
	for i, policy := range policies {
		policyResponses[i] = policy.ToResponse()
	}

	response.Success(c, http.StatusOK, policyResponses)
}

func (h *RBACHandler) GetPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	policy, err := h.rbacService.GetPolicy(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Policy not found", err)
		return
	}

	response.Success(c, http.StatusOK, policy.ToResponse())
}

func (h *RBACHandler) UpdatePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	var input models.UpdatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	policy, err := h.rbacService.UpdatePolicy(id, &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, policy.ToResponse())
}

func (h *RBACHandler) DeletePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	if err := h.rbacService.DeletePolicy(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}

// ============================================================================
// Permissions
// ============================================================================

func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var input models.CreatePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	permission, err := h.rbacService.CreatePermission(&input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, permission.ToResponse())
}

func (h *RBACHandler) ListPermissions(c *gin.Context) {
	service := c.Query("service")
	var permissions []*models.Permission
	var err error

	if service != "" {
		permissions, err = h.rbacService.ListPermissionsByService(service)
	} else {
		permissions, err = h.rbacService.ListPermissions()
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list permissions", err)
		return
	}

	permissionResponses := make([]*models.PermissionResponse, len(permissions))
	for i, permission := range permissions {
		permissionResponses[i] = permission.ToResponse()
	}

	response.Success(c, http.StatusOK, permissionResponses)
}

func (h *RBACHandler) GetPermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	permission, err := h.rbacService.GetPermission(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Permission not found", err)
		return
	}

	response.Success(c, http.StatusOK, permission.ToResponse())
}

func (h *RBACHandler) DeletePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	if err := h.rbacService.DeletePermission(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// ============================================================================
// Policy-Permission assignments
// ============================================================================

func (h *RBACHandler) AssignPermissionsToPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	var input models.AssignPermissionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := h.rbacService.AssignPermissionsToPolicy(id, input.PermissionIDs); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

func (h *RBACHandler) RevokePermissionFromPolicy(c *gin.Context) {
	policyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	if err := h.rbacService.RevokePermissionFromPolicy(policyID, permissionID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Permission revoked successfully"})
}

// ============================================================================
// Authorization
// ============================================================================

func (h *RBACHandler) Authorize(c *gin.Context) {
	tenantIDStr := c.Query("tenant_id")
	userID := c.Query("user_id")
	service := c.Query("service")
	entity := c.Query("entity")
	action := c.Query("action")

	if tenantIDStr == "" || userID == "" || service == "" || entity == "" || action == "" {
		response.Error(c, http.StatusBadRequest, "Missing required query parameters", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tenant_id", err)
		return
	}

	hasPermission, err := h.rbacService.CheckUserPermission(tenantID, userID, service, entity, action)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check permission", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"authorized": hasPermission,
	})
}

func (h *RBACHandler) GetUserPermissions(c *gin.Context) {
	tenantIDStr := c.Query("tenant_id")
	userID := c.Query("user_id")

	if tenantIDStr == "" || userID == "" {
		response.Error(c, http.StatusBadRequest, "Missing tenant_id or user_id", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tenant_id", err)
		return
	}

	permissions, err := h.rbacService.GetUserPermissions(tenantID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user permissions", err)
		return
	}

	permissionResponses := make([]*models.PermissionResponse, len(permissions))
	for i, permission := range permissions {
		permissionResponses[i] = permission.ToResponse()
	}

	response.Success(c, http.StatusOK, permissionResponses)
}

// ============================================================================
// Role-Policy assignments (was Relation-Role)
// ============================================================================

func (h *RBACHandler) AssignPoliciesToRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	var input models.AssignPoliciesToRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := h.rbacService.AssignPoliciesToRole(id, input.PolicyIDs); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Policies assigned successfully"})
}

func (h *RBACHandler) RevokePolicyFromRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	policyID, err := uuid.Parse(c.Param("policy_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	if err := h.rbacService.RevokePolicyFromRole(roleID, policyID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Policy revoked successfully"})
}

func (h *RBACHandler) GetRolePolicies(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	policies, err := h.rbacService.GetRolePolicies(roleID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get role policies", err)
		return
	}

	policyResponses := make([]*models.PolicyResponse, len(policies))
	for i, policy := range policies {
		policyResponses[i] = policy.ToResponse()
	}

	response.Success(c, http.StatusOK, policyResponses)
}
