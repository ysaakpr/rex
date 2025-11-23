package handlers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"go.uber.org/zap"

	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/services"
)

type SystemUserHandler struct {
	systemUserService services.SystemUserService
	logger            *zap.Logger
}

func NewSystemUserHandler(systemUserService services.SystemUserService, logger *zap.Logger) *SystemUserHandler {
	return &SystemUserHandler{
		systemUserService: systemUserService,
		logger:            logger,
	}
}

// CreateSystemUser godoc
// @Summary Create a new system user
// @Description Creates a system user for machine-to-machine authentication (Platform Admin only)
// @Tags system-users
// @Accept json
// @Produce json
// @Param input body models.CreateSystemUserInput true "System user details"
// @Success 201 {object} response.Response{data=models.SystemUserCreateResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /platform/system-users [post]
func (h *SystemUserHandler) CreateSystemUser(c *gin.Context) {
	var input models.CreateSystemUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	// Get current user from session
	sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
	createdBy := sessionContainer.GetUserID()

	h.logger.Info("Creating system user",
		zap.String("name", input.Name),
		zap.String("service_type", input.ServiceType),
		zap.String("created_by", createdBy),
	)

	systemUser, err := h.systemUserService.CreateSystemUser(&input, createdBy)
	if err != nil {
		h.logger.Error("Failed to create system user",
			zap.Error(err),
			zap.String("name", input.Name),
		)
		response.BadRequest(c, err)
		return
	}

	h.logger.Info("System user created successfully",
		zap.String("id", systemUser.ID.String()),
		zap.String("name", systemUser.Name),
		zap.String("email", systemUser.Email),
	)

	response.Created(c, "System user created successfully", systemUser)
}

// GetSystemUser godoc
// @Summary Get system user by ID
// @Description Get details of a specific system user (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param id path string true "System user ID"
// @Success 200 {object} response.Response{data=models.SystemUserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /platform/system-users/{id} [get]
func (h *SystemUserHandler) GetSystemUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	systemUser, err := h.systemUserService.GetSystemUser(id)
	if err != nil {
		response.NotFound(c, "System user not found")
		return
	}

	response.OK(c, systemUser.ToResponse())
}

// ListSystemUsers godoc
// @Summary List all system users
// @Description Get a list of all system users (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param active_only query boolean false "Filter for active users only"
// @Success 200 {object} response.Response{data=[]models.SystemUserResponse}
// @Failure 500 {object} response.Response
// @Router /platform/system-users [get]
func (h *SystemUserHandler) ListSystemUsers(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"

	systemUsers, err := h.systemUserService.ListSystemUsers(activeOnly)
	if err != nil {
		h.logger.Error("Failed to list system users", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}

	responses := make([]*models.SystemUserResponse, len(systemUsers))
	for i, su := range systemUsers {
		responses[i] = su.ToResponse()
	}

	response.OK(c, responses)
}

// UpdateSystemUser godoc
// @Summary Update system user
// @Description Update system user details (Platform Admin only)
// @Tags system-users
// @Accept json
// @Produce json
// @Param id path string true "System user ID"
// @Param input body models.UpdateSystemUserInput true "Update details"
// @Success 200 {object} response.Response{data=models.SystemUserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /platform/system-users/{id} [patch]
func (h *SystemUserHandler) UpdateSystemUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	var input models.UpdateSystemUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	systemUser, err := h.systemUserService.UpdateSystemUser(id, &input)
	if err != nil {
		h.logger.Error("Failed to update system user",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		response.BadRequest(c, err)
		return
	}

	response.OK(c, systemUser.ToResponse())
}

// RegeneratePassword godoc
// @Summary Regenerate system user password
// @Description Generate a new password for a system user and revoke all existing sessions (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param id path string true "System user ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /platform/system-users/{id}/regenerate-password [post]
func (h *SystemUserHandler) RegeneratePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	password, err := h.systemUserService.RegeneratePassword(id)
	if err != nil {
		h.logger.Error("Failed to regenerate password",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		response.BadRequest(c, err)
		return
	}

	// Get updated system user details
	systemUser, err := h.systemUserService.GetSystemUser(id)
	if err != nil {
		h.logger.Error("Failed to get system user after password regeneration",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		response.InternalServerError(c, err)
		return
	}

	h.logger.Info("Password regenerated successfully", zap.String("id", id.String()))

	// Return full user info with new password
	result := map[string]interface{}{
		"password":     password,
		"email":        systemUser.Email,
		"user_id":      systemUser.UserID,
		"service_type": systemUser.ServiceType,
		"message":      "Save this password securely. It will not be shown again. All existing sessions have been revoked.",
	}

	response.OK(c, result)
}

// DeactivateSystemUser godoc
// @Summary Deactivate system user
// @Description Deactivate a system user and revoke all sessions (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param id path string true "System user ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /platform/system-users/{id} [delete]
func (h *SystemUserHandler) DeactivateSystemUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	if err := h.systemUserService.DeactivateSystemUser(id); err != nil {
		h.logger.Error("Failed to deactivate system user",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		response.BadRequest(c, err)
		return
	}

	h.logger.Info("System user deactivated successfully", zap.String("id", id.String()))

	response.OK(c, map[string]string{
		"message": "System user deactivated successfully. All sessions have been revoked.",
	})
}

// RotateWithGracePeriod godoc
// @Summary Rotate credentials with grace period
// @Description Create new credential while keeping old one active for a grace period (Platform Admin only)
// @Tags system-users
// @Accept json
// @Produce json
// @Param id path string true "System user ID"
// @Param input body map[string]int false "Grace period configuration"
// @Success 200 {object} response.Response{data=models.SystemUserCreateResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /platform/system-users/{id}/rotate [post]
func (h *SystemUserHandler) RotateWithGracePeriod(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	// Parse input for grace period (optional)
	var input struct {
		GracePeriodDays int `json:"grace_period_days"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		// If no body provided, use default
		input.GracePeriodDays = 7
	}

	// Default to 7 days if not specified or invalid
	if input.GracePeriodDays <= 0 {
		input.GracePeriodDays = 7
	}

	result, err := h.systemUserService.RotateWithGracePeriod(id, input.GracePeriodDays)
	if err != nil {
		h.logger.Error("Failed to rotate credentials",
			zap.Error(err),
			zap.String("id", id.String()),
			zap.Int("grace_period_days", input.GracePeriodDays),
		)
		response.BadRequest(c, err)
		return
	}

	h.logger.Info("Credentials rotated successfully",
		zap.String("id", id.String()),
		zap.String("new_email", result.Email),
		zap.Int("grace_period_days", input.GracePeriodDays),
	)

	response.OK(c, result)
}

// GetApplicationCredentials godoc
// @Summary Get all credentials for an application
// @Description Get list of all credentials (active and expiring) for a logical application (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param application_name path string true "Application name"
// @Success 200 {object} response.Response{data=[]models.SystemUserResponse}
// @Failure 400 {object} response.Response
// @Router /platform/applications/{application_name}/credentials [get]
func (h *SystemUserHandler) GetApplicationCredentials(c *gin.Context) {
	applicationName := c.Param("application_name")
	if applicationName == "" {
		response.BadRequest(c, errors.New("application_name is required"))
		return
	}

	credentials, err := h.systemUserService.GetSystemUsersByApplication(applicationName)
	if err != nil {
		h.logger.Error("Failed to get application credentials",
			zap.Error(err),
			zap.String("application_name", applicationName),
		)
		response.InternalServerError(c, err)
		return
	}

	// Convert to responses
	responses := make([]*models.SystemUserResponse, len(credentials))
	for i, cred := range credentials {
		responses[i] = cred.ToResponse()
	}

	response.OK(c, map[string]interface{}{
		"application_name": applicationName,
		"credentials":      responses,
	})
}

// RevokeOldCredentials godoc
// @Summary Revoke all non-primary credentials
// @Description Immediately deactivate all non-primary credentials for an application (Platform Admin only)
// @Tags system-users
// @Produce json
// @Param application_name path string true "Application name"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 400 {object} response.Response
// @Router /platform/applications/{application_name}/revoke-old [post]
func (h *SystemUserHandler) RevokeOldCredentials(c *gin.Context) {
	applicationName := c.Param("application_name")
	if applicationName == "" {
		response.BadRequest(c, errors.New("application_name is required"))
		return
	}

	count, err := h.systemUserService.RevokeNonPrimary(applicationName)
	if err != nil {
		h.logger.Error("Failed to revoke old credentials",
			zap.Error(err),
			zap.String("application_name", applicationName),
		)
		response.InternalServerError(c, err)
		return
	}

	h.logger.Info("Old credentials revoked",
		zap.String("application_name", applicationName),
		zap.Int("count", count),
	)

	response.OK(c, map[string]interface{}{
		"revoked_count": count,
		"message":       fmt.Sprintf("Successfully revoked %d old credential(s)", count),
	})
}
