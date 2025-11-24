package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/api/middleware"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/services"
	"gorm.io/gorm"
)

type TenantHandler struct {
	tenantService services.TenantService
	db            *gorm.DB
}

func NewTenantHandler(tenantService services.TenantService, db *gorm.DB) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		db:            db,
	}
}

// CreateTenant godoc
// @Summary Create a new tenant (self-onboarding)
// @Tags tenants
// @Accept json
// @Produce json
// @Param input body models.CreateTenantInput true "Tenant creation input"
// @Success 201 {object} response.Response{data=models.TenantResponse}
// @Router /tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input models.CreateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	tenant, err := h.tenantService.CreateTenant(&input, userID)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	tenantResp := tenant.ToResponse()
	// Creator is automatically added as a member
	tenantResp.MemberCount = 1

	response.Created(c, "Tenant created successfully", tenantResp)
}

// CreateManagedTenant godoc
// @Summary Create a managed tenant (super admin only)
// @Tags tenants
// @Accept json
// @Produce json
// @Param input body object true "Managed tenant creation input"
// @Success 201 {object} response.Response{data=models.TenantResponse}
// @Router /tenants/managed [post]
func (h *TenantHandler) CreateManagedTenant(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input struct {
		models.CreateTenantInput
		AdminEmail string `json:"admin_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	tenant, err := h.tenantService.CreateManagedTenant(&input.CreateTenantInput, input.AdminEmail, userID)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	tenantResp := tenant.ToResponse()
	// Admin will be added as a member when they accept the invitation
	tenantResp.MemberCount = 0

	response.Created(c, "Managed tenant created successfully", tenantResp)
}

// GetTenant godoc
// @Summary Get tenant by ID
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Router /tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	tenantResp := tenant.ToResponse()

	// Count active members for this tenant
	var memberCount int64
	h.db.Table("tenant_members").
		Where("tenant_id = ?", tenant.ID).
		Where("status = ?", "active").
		Count(&memberCount)

	tenantResp.MemberCount = int(memberCount)

	response.OK(c, tenantResp)
}

// GetTenantForPlatformAdmin godoc
// @Summary Get tenant by ID (platform admins only, no membership required)
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Router /platform/tenants/{id} [get]
func (h *TenantHandler) GetTenantForPlatformAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	tenantResp := tenant.ToResponse()

	// Count active members for this tenant
	var memberCount int64
	h.db.Table("tenant_members").
		Where("tenant_id = ?", tenant.ID).
		Where("status = ?", "active").
		Count(&memberCount)

	tenantResp.MemberCount = int(memberCount)

	response.OK(c, tenantResp)
}

// ListAllTenants godoc
// @Summary List all tenants (platform admins only)
// @Tags tenants
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response{data=models.PaginatedResponse}
// @Router /platform/tenants [get]
func (h *TenantHandler) ListAllTenants(c *gin.Context) {
	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, err)
		return
	}

	tenants, total, err := h.tenantService.GetAllTenants(&pagination)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Convert to response format and add member counts
	tenantResponses := make([]*models.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResp := tenant.ToResponse()

		// Count active members for this tenant
		var memberCount int64
		h.db.Table("tenant_members").
			Where("tenant_id = ?", tenant.ID).
			Where("status = ?", "active").
			Count(&memberCount)

		tenantResp.MemberCount = int(memberCount)
		tenantResponses[i] = tenantResp
	}

	pagination.Normalize()
	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}

	result := models.PaginatedResponse{
		Data:       tenantResponses,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	response.OK(c, result)
}

// ListTenants godoc
// @Summary List user's tenants
// @Tags tenants
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response{data=models.PaginatedResponse}
// @Router /tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, err)
		return
	}

	tenants, total, err := h.tenantService.GetUserTenants(userID, &pagination)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Convert to response format and add member counts
	tenantResponses := make([]*models.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResp := tenant.ToResponse()

		// Count active members for this tenant
		var memberCount int64
		h.db.Table("tenant_members").
			Where("tenant_id = ?", tenant.ID).
			Where("status = ?", "active").
			Count(&memberCount)

		tenantResp.MemberCount = int(memberCount)
		tenantResponses[i] = tenantResp
	}

	pagination.Normalize()
	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}

	result := models.PaginatedResponse{
		Data:       tenantResponses,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	response.OK(c, result)
}

// UpdateTenant godoc
// @Summary Update tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param input body models.UpdateTenantInput true "Tenant update input"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Router /tenants/{id} [patch]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	var input models.UpdateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	tenant, err := h.tenantService.UpdateTenant(id, &input)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	tenantResp := tenant.ToResponse()

	// Count active members for this tenant
	var memberCount int64
	h.db.Table("tenant_members").
		Where("tenant_id = ?", tenant.ID).
		Where("status = ?", "active").
		Count(&memberCount)

	tenantResp.MemberCount = int(memberCount)

	response.OK(c, tenantResp)
}

// DeleteTenant godoc
// @Summary Delete tenant
// @Tags tenants
// @Param id path string true "Tenant ID"
// @Success 204
// @Router /tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	if err := h.tenantService.DeleteTenant(id); err != nil {
		response.BadRequest(c, err)
		return
	}

	response.NoContent(c)
}

// GetTenantStatus godoc
// @Summary Get tenant initialization status
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} response.Response{data=object}
// @Router /tenants/{id}/status [get]
func (h *TenantHandler) GetTenantStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	status, err := h.tenantService.GetTenantStatus(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"status": status,
	})
}
