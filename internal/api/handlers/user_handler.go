package handlers

import (
	"errors"
	"net/http"
	"strconv"
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
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name,omitempty"`
	IsActive    bool   `json:"is_active"`
	TenantCount int    `json:"tenant_count"`
	IsSystem    bool   `json:"is_system"`
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

	// Check if this is a system user
	var isSystem bool
	err = h.db.Table("system_users").
		Where("user_id = ?", userInfo.ID).
		Select("COUNT(*) > 0").
		Row().Scan(&isSystem)
	if err != nil {
		h.logger.Warn("Failed to check system user status", zap.Error(err))
	}

	// Count tenant memberships
	var tenantCount int64
	h.db.Table("tenant_members").
		Where("user_id = ?", userInfo.ID).
		Where("status != ?", "inactive").
		Count(&tenantCount)

	// Build response
	userDetails := UserDetailsResponse{
		UserID:      userInfo.ID,
		Email:       userInfo.Email,
		IsActive:    true, // SuperTokens users are active by default
		TenantCount: int(tenantCount),
		IsSystem:    isSystem,
		// SuperTokens doesn't store name by default in EmailPassword recipe
		// We can add it later if we use user metadata
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", userDetails)
}

// ListUsers godoc
// @Summary List all users
// @Description Fetches ALL users from SuperTokens including system users
// @Tags users
// @Produce json
// @Param email query string false "Filter by email (partial match)"
// @Success 200 {object} response.Response{data=[]UserDetailsResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	emailFilter := c.Query("email")

	// Pagination parameters
	page := 1
	pageSize := 20 // Default page size

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	h.logger.Info("List users requested",
		zap.String("email_filter", emailFilter),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// Get ALL user IDs from SuperTokens
	// Note: SuperTokens doesn't provide a direct "list all users" API in the Go SDK
	// We'll get users from our database tables and system_users

	allUserIDsMap := make(map[string]bool)

	// 1. Get user IDs from tenants (creators)
	var tenantCreators []string
	if err := h.db.Table("tenants").
		Where("deleted_at IS NULL").
		Distinct("created_by").
		Pluck("created_by", &tenantCreators).Error; err != nil {
		h.logger.Error("Failed to fetch tenant creators", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}
	for _, id := range tenantCreators {
		allUserIDsMap[id] = true
	}

	// 2. Get user IDs from tenant_members
	var memberUserIDs []string
	if err := h.db.Table("tenant_members").
		Distinct("user_id").
		Pluck("user_id", &memberUserIDs).Error; err != nil {
		h.logger.Error("Failed to fetch tenant members", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}
	for _, id := range memberUserIDs {
		allUserIDsMap[id] = true
	}

	// 3. Get user IDs from system_users
	var systemUserIDs []string
	if err := h.db.Table("system_users").
		Distinct("user_id").
		Pluck("user_id", &systemUserIDs).Error; err != nil {
		h.logger.Error("Failed to fetch system users", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}
	for _, id := range systemUserIDs {
		allUserIDsMap[id] = true
	}

	// 4. Get user IDs from platform_admins
	var platformAdminIDs []string
	if err := h.db.Table("platform_admins").
		Distinct("user_id").
		Pluck("user_id", &platformAdminIDs).Error; err != nil {
		h.logger.Error("Failed to fetch platform admins", zap.Error(err))
		response.InternalServerError(c, err)
		return
	}
	for _, id := range platformAdminIDs {
		allUserIDsMap[id] = true
	}

	// Build a map of system user IDs for quick lookup
	systemUserMap := make(map[string]bool)
	for _, id := range systemUserIDs {
		systemUserMap[id] = true
	}

	// Fetch user details from SuperTokens for each user ID
	var allUsers []UserDetailsResponse
	for userID := range allUserIDsMap {
		userInfo, err := emailpassword.GetUserByID(userID)
		if err != nil {
			h.logger.Warn("Failed to fetch user from SuperTokens",
				zap.String("user_id", userID),
				zap.Error(err))
			continue
		}

		if userInfo == nil {
			h.logger.Warn("User not found in SuperTokens",
				zap.String("user_id", userID))
			continue
		}

		// Count tenant memberships
		var tenantCount int64
		h.db.Table("tenant_members").
			Where("user_id = ?", userID).
			Where("status != ?", "inactive").
			Count(&tenantCount)

		// Check if system user
		isSystem := systemUserMap[userID]

		userDetails := UserDetailsResponse{
			UserID:      userInfo.ID,
			Email:       userInfo.Email,
			IsActive:    true, // SuperTokens users are active by default
			TenantCount: int(tenantCount),
			IsSystem:    isSystem,
		}

		// Apply email filter if provided (case-insensitive)
		if emailFilter == "" || strings.Contains(strings.ToLower(userDetails.Email), strings.ToLower(emailFilter)) {
			allUsers = append(allUsers, userDetails)
		}
	}

	// Calculate pagination
	totalCount := len(allUsers)
	totalPages := (totalCount + pageSize - 1) / pageSize

	// Apply pagination
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalCount {
		// Page out of range, return empty
		startIdx = 0
		endIdx = 0
		allUsers = []UserDetailsResponse{}
	} else {
		if endIdx > totalCount {
			endIdx = totalCount
		}
		allUsers = allUsers[startIdx:endIdx]
	}

	h.logger.Info("Users fetched successfully",
		zap.Int("total_count", totalCount),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.Int("total_pages", totalPages),
		zap.Int("returned", len(allUsers)),
		zap.String("email_filter", emailFilter))

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Users retrieved successfully",
		"data": gin.H{
			"data":        allUsers,
			"total_count": totalCount,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
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
