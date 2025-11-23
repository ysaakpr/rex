package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/ysaakpr/rex/internal/pkg/response"
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
	ID          string `json:"id"`      // Alias for UserID for convenience
	UserID      string `json:"user_id"` // Original field for backward compatibility
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
		ID:          userInfo.ID,
		UserID:      userInfo.ID,
		Email:       userInfo.Email,
		Name:        userInfo.Email, // Use email as name fallback
		IsActive:    true,           // SuperTokens users are active by default
		TenantCount: int(tenantCount),
		IsSystem:    isSystem,
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
			ID:          userInfo.ID,
			UserID:      userInfo.ID,
			Email:       userInfo.Email,
			Name:        userInfo.Email, // Use email as name fallback
			IsActive:    true,           // SuperTokens users are active by default
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

// UserTenantMembership represents a user's membership in a tenant
type UserTenantMembership struct {
	TenantID   string `gorm:"column:tenant_id" json:"tenant_id"`
	TenantName string `gorm:"column:tenant_name" json:"tenant_name"`
	RoleID     string `gorm:"column:role_id" json:"role_id"`
	RoleName   string `gorm:"column:role_name" json:"role_name"`
	Status     string `gorm:"column:status" json:"status"`
	JoinedAt   string `gorm:"column:joined_at" json:"joined_at"`
}

// GetUserTenants godoc
// @Summary Get user's tenant memberships
// @Description Fetches all tenants a user belongs to with their relations and roles
// @Tags users
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=[]UserTenantMembership}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{user_id}/tenants [get]
func (h *UserHandler) GetUserTenants(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, errors.New("user_id is required"))
		return
	}

	h.logger.Info("Get user tenants requested", zap.String("user_id", userID))

	// Query tenant memberships with joins to get tenant and role names
	var memberships []UserTenantMembership
	err := h.db.Table("tenant_members").
		Select(`
			tenant_members.tenant_id,
			tenants.name as tenant_name,
			tenant_members.role_id,
			roles.name as role_name,
			tenant_members.status,
			tenant_members.joined_at
		`).
		Joins("LEFT JOIN tenants ON tenants.id = tenant_members.tenant_id").
		Joins("LEFT JOIN roles ON roles.id = tenant_members.role_id").
		Where("tenant_members.user_id = ?", userID).
		Where("tenant_members.status = ?", "active").
		Order("tenant_members.joined_at DESC").
		Scan(&memberships).Error

	if err != nil {
		h.logger.Error("Failed to fetch user tenants",
			zap.String("user_id", userID),
			zap.Error(err))
		response.InternalServerError(c, err)
		return
	}

	h.logger.Info("User tenants retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(memberships)))

	response.Success(c, http.StatusOK, "User tenants retrieved successfully", memberships)
}

// SearchUsers godoc
// @Summary Search users by name or email
// @Description Search for users by partial name or email match, returns simplified results
// @Tags users
// @Produce json
// @Param q query string true "Search query (name or email)"
// @Success 200 {object} response.Response{data=[]UserDetailsResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" || len(query) < 2 {
		response.BadRequest(c, errors.New("search query must be at least 2 characters"))
		return
	}

	h.logger.Info("Search users requested", zap.String("query", query))

	// Collect all user IDs from various sources
	allUserIDsMap := make(map[string]bool)

	// 1. Get user IDs from tenants (creators)
	var tenantCreators []string
	if err := h.db.Table("tenants").
		Where("deleted_at IS NULL").
		Distinct("created_by").
		Pluck("created_by", &tenantCreators).Error; err == nil {
		for _, id := range tenantCreators {
			allUserIDsMap[id] = true
		}
	}

	// 2. Get user IDs from tenant_members
	var memberUserIDs []string
	if err := h.db.Table("tenant_members").
		Distinct("user_id").
		Pluck("user_id", &memberUserIDs).Error; err == nil {
		for _, id := range memberUserIDs {
			allUserIDsMap[id] = true
		}
	}

	// 3. Get user IDs from system_users
	var systemUserIDs []string
	if err := h.db.Table("system_users").
		Distinct("user_id").
		Pluck("user_id", &systemUserIDs).Error; err == nil {
		for _, id := range systemUserIDs {
			allUserIDsMap[id] = true
		}
	}

	// 4. Get user IDs from platform_admins
	var platformAdminIDs []string
	if err := h.db.Table("platform_admins").
		Distinct("user_id").
		Pluck("user_id", &platformAdminIDs).Error; err == nil {
		for _, id := range platformAdminIDs {
			allUserIDsMap[id] = true
		}
	}

	// Build a map of system user IDs for quick lookup
	systemUserMap := make(map[string]bool)
	for _, id := range systemUserIDs {
		systemUserMap[id] = true
	}

	// Fetch user details from SuperTokens and filter
	queryLower := strings.ToLower(query)
	var matchedUsers []UserDetailsResponse
	for userID := range allUserIDsMap {
		userInfo, err := emailpassword.GetUserByID(userID)
		if err != nil || userInfo == nil {
			continue
		}

		// Check if email matches the search query
		if strings.Contains(strings.ToLower(userInfo.Email), queryLower) {
			// Count tenant memberships
			var tenantCount int64
			h.db.Table("tenant_members").
				Where("user_id = ?", userID).
				Where("status != ?", "inactive").
				Count(&tenantCount)

			matchedUsers = append(matchedUsers, UserDetailsResponse{
				ID:          userInfo.ID,
				UserID:      userInfo.ID,
				Email:       userInfo.Email,
				Name:        userInfo.Email, // Use email as name fallback
				IsActive:    true,
				TenantCount: int(tenantCount),
				IsSystem:    systemUserMap[userID],
			})

			// Limit results to prevent overwhelming the UI
			if len(matchedUsers) >= 20 {
				break
			}
		}
	}

	h.logger.Info("User search completed",
		zap.String("query", query),
		zap.Int("matches", len(matchedUsers)))

	response.Success(c, http.StatusOK, "Users retrieved successfully", matchedUsers)
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
				ID:     userInfo.ID,
				UserID: userInfo.ID,
				Email:  userInfo.Email,
				Name:   userInfo.Email, // Use email as name fallback
			}
		}
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", usersMap)
}
