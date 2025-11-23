package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/vyshakhp/utm-backend/internal/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewUserHandler(logger *zap.Logger, db *gorm.DB) *UserHandler {
	return &UserHandler{
		logger: logger,
		db:     db,
	}
}

// UserDetailsResponse defines the response structure for user details
type UserDetailsResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name,omitempty"`
}

// GetUserDetails godoc
// @Summary Get user details by user ID
// @Description Fetches basic user information (email, name) from SuperTokens by user ID
// @Tags users
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=UserDetailsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{user_id} [get]
func (h *UserHandler) GetUserDetails(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, errors.New("user_id is required"))
		return
	}

	// Fetch user from SuperTokens
	userInfo, err := emailpassword.GetUserByID(userID)
	if err != nil {
		h.logger.Error("Failed to fetch user from SuperTokens",
			zap.String("user_id", userID),
			zap.Error(err))
		response.InternalServerError(c, err)
		return
	}

	if userInfo == nil {
		response.NotFound(c, "User not found")
		return
	}

	// Build response
	userDetails := UserDetailsResponse{
		UserID: userInfo.ID,
		Email:  userInfo.Email,
		// SuperTokens doesn't store name by default in EmailPassword recipe
		// We can add it later if we use user metadata
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", userDetails)
}

// ListUsers godoc
// @Summary List all users
// @Description Fetches all users that are in the system (tenant owners and members)
// @Tags users
// @Produce json
// @Param email query string false "Filter by email (partial match)"
// @Success 200 {object} response.Response{data=[]UserDetailsResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	emailFilter := c.Query("email")

	h.logger.Info("List users requested",
		zap.String("email_filter", emailFilter))

	// Get unique user IDs from tenants (creators) and tenant_members
	var userIDs []string

	// Get user IDs from tenants (created_by = tenant owner)
	if err := h.db.Table("tenants").
		Where("deleted_at IS NULL").
		Distinct("created_by").
		Pluck("created_by", &userIDs).Error; err != nil {
		h.logger.Error("Failed to fetch tenant creators", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}

	// Get user IDs from tenant_members (only active members)
	var memberUserIDs []string
	if err := h.db.Table("tenant_members").
		Where("status != ?", "inactive").
		Distinct("user_id").
		Pluck("user_id", &memberUserIDs).Error; err != nil {
		h.logger.Error("Failed to fetch tenant members", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}

	// Combine and deduplicate user IDs
	userIDMap := make(map[string]bool)
	for _, id := range userIDs {
		userIDMap[id] = true
	}
	for _, id := range memberUserIDs {
		userIDMap[id] = true
	}

	// Fetch user details from SuperTokens for each user ID
	var allUsers []UserDetailsResponse
	for userID := range userIDMap {
		userInfo, err := emailpassword.GetUserByID(userID)
		if err != nil {
			h.logger.Warn("Failed to fetch user from SuperTokens",
				zap.String("user_id", userID),
				zap.Error(err))
			continue
		}

		if userInfo == nil {
			continue
		}

		userDetails := UserDetailsResponse{
			UserID: userInfo.ID,
			Email:  userInfo.Email,
		}

		// Apply email filter if provided (case-insensitive)
		if emailFilter == "" || strings.Contains(strings.ToLower(userDetails.Email), strings.ToLower(emailFilter)) {
			allUsers = append(allUsers, userDetails)
		}
	}

	h.logger.Info("Users fetched successfully",
		zap.Int("total_count", len(allUsers)),
		zap.String("email_filter", emailFilter))

	response.Success(c, http.StatusOK, "Users retrieved successfully", allUsers)
}

// GetUserTenants godoc
// @Summary Get user's tenant memberships
// @Description Fetches all tenants a user belongs to with their relations and roles
// @Tags users
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=[]map[string]interface{}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{user_id}/tenants [get]
func (h *UserHandler) GetUserTenants(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, errors.New("user_id is required"))
		return
	}

	// This would query the database for all tenant_members records for this user
	// Including tenant info, relation, and roles
	// Implementation would go here

	h.logger.Info("Get user tenants requested", zap.String("user_id", userID))

	response.Success(c, http.StatusOK, "User tenants retrieved", []map[string]interface{}{})
}

// GetBatchUserDetails godoc
// @Summary Get details for multiple users
// @Description Fetches basic user information for multiple user IDs in a single request
// @Tags users
// @Accept json
// @Produce json
// @Param user_ids body []string true "Array of user IDs"
// @Success 200 {object} response.Response{data=map[string]UserDetailsResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/batch [post]
func (h *UserHandler) GetBatchUserDetails(c *gin.Context) {
	var userIDs []string
	if err := c.ShouldBindJSON(&userIDs); err != nil {
		response.BadRequest(c, err)
		return
	}

	if len(userIDs) == 0 {
		response.BadRequest(c, errors.New("user_ids array is required and cannot be empty"))
		return
	}

	// Limit batch size to prevent abuse
	if len(userIDs) > 100 {
		response.BadRequest(c, errors.New("maximum 100 user IDs allowed per request"))
		return
	}

	// Fetch users from SuperTokens
	usersMap := make(map[string]UserDetailsResponse)
	for _, userID := range userIDs {
		userInfo, err := emailpassword.GetUserByID(userID)
		if err != nil {
			h.logger.Warn("Failed to fetch user from SuperTokens",
				zap.String("user_id", userID),
				zap.Error(err))
			continue // Skip this user but continue with others
		}

		if userInfo != nil {
			usersMap[userID] = UserDetailsResponse{
				UserID: userInfo.ID,
				Email:  userInfo.Email,
			}
		}
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", usersMap)
}
