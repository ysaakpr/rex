package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/api/middleware"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/services"
)

type PlatformAdminHandler struct {
	adminService services.PlatformAdminService
}

func NewPlatformAdminHandler(adminService services.PlatformAdminService) *PlatformAdminHandler {
	return &PlatformAdminHandler{
		adminService: adminService,
	}
}

// CreateAdmin godoc
// @Summary Create a platform admin
// @Tags platform
// @Accept json
// @Produce json
// @Param input body models.CreatePlatformAdminInput true "Admin input"
// @Success 201 {object} response.Response{data=models.PlatformAdminResponse}
// @Router /platform/admins [post]
func (h *PlatformAdminHandler) CreateAdmin(c *gin.Context) {
	currentUserID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input models.CreatePlatformAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	admin, err := h.adminService.CreateAdmin(input.UserID, currentUserID)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.Created(c, "Platform admin created successfully", admin.ToResponse())
}

// ListAdmins godoc
// @Summary List all platform admins
// @Tags platform
// @Produce json
// @Success 200 {object} response.Response{data=[]models.PlatformAdminResponse}
// @Router /platform/admins [get]
func (h *PlatformAdminHandler) ListAdmins(c *gin.Context) {
	admins, err := h.adminService.ListAdmins()
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	adminResponses := make([]*models.PlatformAdminResponse, len(admins))
	for i, admin := range admins {
		adminResponses[i] = admin.ToResponse()
	}

	response.Success(c, 200, "Platform admins retrieved successfully", adminResponses)
}

// GetAdmin godoc
// @Summary Get a platform admin
// @Tags platform
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=models.PlatformAdminResponse}
// @Router /platform/admins/{user_id} [get]
func (h *PlatformAdminHandler) GetAdmin(c *gin.Context) {
	userID := c.Param("user_id")

	admin, err := h.adminService.GetAdmin(userID)
	if err != nil {
		response.NotFound(c, "Platform admin not found")
		return
	}

	response.Success(c, 200, "Platform admin retrieved successfully", admin.ToResponse())
}

// DeleteAdmin godoc
// @Summary Delete a platform admin
// @Tags platform
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response
// @Router /platform/admins/{user_id} [delete]
func (h *PlatformAdminHandler) DeleteAdmin(c *gin.Context) {
	userID := c.Param("user_id")

	if err := h.adminService.DeleteAdmin(userID); err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, 200, "Platform admin deleted successfully", nil)
}

// CheckPlatformAdmin godoc
// @Summary Check if current user is platform admin
// @Tags platform
// @Produce json
// @Success 200 {object} response.Response{data=map[string]bool}
// @Router /platform/admins/check [get]
func (h *PlatformAdminHandler) CheckPlatformAdmin(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isAdmin, err := h.adminService.IsPlatformAdmin(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, 200, "Platform admin status checked", map[string]bool{
		"is_platform_admin": isAdmin,
	})
}
