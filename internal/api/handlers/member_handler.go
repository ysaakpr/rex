package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vyshakhp/utm-backend/internal/api/middleware"
	"github.com/vyshakhp/utm-backend/internal/models"
	"github.com/vyshakhp/utm-backend/internal/pkg/response"
	"github.com/vyshakhp/utm-backend/internal/services"
)

type MemberHandler struct {
	memberService services.MemberService
}

func NewMemberHandler(memberService services.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

// AddMember godoc
// @Summary Add a member to tenant
// @Tags members
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param input body models.AddMemberInput true "Member input"
// @Success 201 {object} response.Response{data=models.MemberResponse}
// @Router /tenants/{tenant_id}/members [post]
func (h *MemberHandler) AddMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input models.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	member, err := h.memberService.AddMember(tenantID, &input, userID)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.Created(c, "Member added successfully", member.ToResponse())
}

// ListMembers godoc
// @Summary List tenant members
// @Tags members
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response{data=models.PaginatedResponse}
// @Router /tenants/{tenant_id}/members [get]
func (h *MemberHandler) ListMembers(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, err)
		return
	}

	members, total, err := h.memberService.GetTenantMembers(tenantID, &pagination)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Convert to response format
	memberResponses := make([]*models.MemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = member.ToResponse()
	}

	pagination.Normalize()
	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}

	result := models.PaginatedResponse{
		Data:       memberResponses,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	response.OK(c, result)
}

// GetMember godoc
// @Summary Get member details
// @Tags members
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=models.MemberResponse}
// @Router /tenants/{tenant_id}/members/{user_id} [get]
func (h *MemberHandler) GetMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID := c.Param("user_id")

	member, err := h.memberService.GetMember(tenantID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, member.ToResponse())
}

// UpdateMember godoc
// @Summary Update member
// @Tags members
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Param input body models.UpdateMemberInput true "Member update input"
// @Success 200 {object} response.Response{data=models.MemberResponse}
// @Router /tenants/{tenant_id}/members/{user_id} [patch]
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID := c.Param("user_id")

	// Get member ID first
	member, err := h.memberService.GetMember(tenantID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var input models.UpdateMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	updatedMember, err := h.memberService.UpdateMember(member.ID, &input)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, updatedMember.ToResponse())
}

// RemoveMember godoc
// @Summary Remove member from tenant
// @Tags members
// @Param tenant_id path string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Success 204
// @Router /tenants/{tenant_id}/members/{user_id} [delete]
func (h *MemberHandler) RemoveMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID := c.Param("user_id")

	// Get member ID first
	member, err := h.memberService.GetMember(tenantID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	if err := h.memberService.RemoveMember(member.ID); err != nil {
		response.BadRequest(c, err)
		return
	}

	response.NoContent(c)
}

// AssignRoles godoc
// @Summary Assign roles to member
// @Tags members
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Param input body object true "Role IDs"
// @Success 200 {object} response.Response
// @Router /tenants/{tenant_id}/members/{user_id}/roles [post]
func (h *MemberHandler) AssignRoles(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID := c.Param("user_id")

	// Get member ID first
	member, err := h.memberService.GetMember(tenantID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var input struct {
		RoleIDs []uuid.UUID `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	if err := h.memberService.AssignRolesToMember(member.ID, input.RoleIDs); err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, gin.H{"message": "Roles assigned successfully"})
}

// RemoveRole godoc
// @Summary Remove role from member
// @Tags members
// @Param tenant_id path string true "Tenant ID"
// @Param user_id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 204
// @Router /tenants/{tenant_id}/members/{user_id}/roles/{role_id} [delete]
func (h *MemberHandler) RemoveRole(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID := c.Param("user_id")
	roleIDStr := c.Param("role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	// Get member ID first
	member, err := h.memberService.GetMember(tenantID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	if err := h.memberService.RemoveRoleFromMember(member.ID, roleID); err != nil {
		response.BadRequest(c, err)
		return
	}

	response.NoContent(c)
}
