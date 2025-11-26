package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/api/middleware"
	"github.com/ysaakpr/rex/internal/config"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/services"
)

type InvitationHandler struct {
	invitationService services.InvitationService
	cfg               *config.Config
}

func NewInvitationHandler(invitationService services.InvitationService, cfg *config.Config) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		cfg:               cfg,
	}
}

// buildInvitationURL generates the full invitation URL using the configured base URL
func (h *InvitationHandler) buildInvitationURL(token string) string {
	return fmt.Sprintf("%s?token=%s", h.cfg.Invitation.BaseURL, token)
}

// enrichInvitationResponse adds the invitation URL to the response
func (h *InvitationHandler) enrichInvitationResponse(inv *models.UserInvitation) *models.InvitationResponse {
	resp := inv.ToResponse()
	resp.InvitationURL = h.buildInvitationURL(inv.Token)
	return resp
}

// CreateInvitation godoc
// @Summary Invite a user to tenant
// @Tags invitations
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param input body models.CreateInvitationInput true "Invitation input"
// @Success 201 {object} response.Response{data=models.InvitationResponse}
// @Router /tenants/{tenant_id}/invitations [post]
func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
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

	var input models.CreateInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	invitation, err := h.invitationService.CreateInvitation(tenantID, &input, userID)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.Created(c, "Invitation sent successfully", h.enrichInvitationResponse(invitation))
}

// ListInvitations godoc
// @Summary List tenant invitations
// @Tags invitations
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response{data=models.PaginatedResponse}
// @Router /tenants/{tenant_id}/invitations [get]
func (h *InvitationHandler) ListInvitations(c *gin.Context) {
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

	invitations, total, err := h.invitationService.GetTenantInvitations(tenantID, &pagination)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Convert to response format with invitation URLs
	invitationResponses := make([]*models.InvitationResponse, len(invitations))
	for i, invitation := range invitations {
		invitationResponses[i] = h.enrichInvitationResponse(invitation)
	}

	pagination.Normalize()
	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}

	result := models.PaginatedResponse{
		Data:       invitationResponses,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	response.OK(c, result)
}

// GetInvitationByToken godoc
// @Summary Get invitation details by token
// @Tags invitations
// @Produce json
// @Param token path string true "Invitation Token"
// @Success 200 {object} response.Response{data=models.InvitationResponse}
// @Router /invitations/{token} [get]
func (h *InvitationHandler) GetInvitationByToken(c *gin.Context) {
	token := c.Param("token")

	invitation, err := h.invitationService.GetInvitationByToken(token)
	if err != nil {
		errMsg := err.Error()
		// Return 400 for invalid invitation states (already accepted, expired, cancelled)
		// Return 404 only for truly not found
		if errMsg == "invitation not found" {
			response.NotFound(c, errMsg)
		} else {
			response.BadRequest(c, err)
		}
		return
	}

	response.OK(c, h.enrichInvitationResponse(invitation))
}

// AcceptInvitation godoc
// @Summary Accept invitation
// @Tags invitations
// @Produce json
// @Param token path string true "Invitation Token"
// @Success 200 {object} response.Response{data=models.MemberResponse}
// @Router /invitations/{token}/accept [post]
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get user email to validate against invitation
	userEmail, err := middleware.GetUserEmail(c)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	member, err := h.invitationService.AcceptInvitation(token, userID, userEmail)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, member.ToResponse())
}

// CheckPendingInvitations godoc
// @Summary Check and auto-accept pending invitations for current user
// @Tags invitations
// @Produce json
// @Success 200 {object} response.Response{data=[]models.MemberResponse}
// @Router /invitations/check-pending [post]
func (h *InvitationHandler) CheckPendingInvitations(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get user email from SuperTokens
	email, err := middleware.GetUserEmail(c)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	members, err := h.invitationService.CheckAndAcceptPendingInvitations(email, userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Convert to response format
	memberResponses := make([]*models.MemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = member.ToResponse()
	}

	response.OK(c, gin.H{
		"accepted_count": len(members),
		"memberships":    memberResponses,
	})
}

// CancelInvitation godoc
// @Summary Cancel invitation
// @Tags invitations
// @Param id path string true "Invitation ID"
// @Success 204
// @Router /invitations/{id} [delete]
func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	if err := h.invitationService.CancelInvitation(id); err != nil {
		response.BadRequest(c, err)
		return
	}

	response.NoContent(c)
}
