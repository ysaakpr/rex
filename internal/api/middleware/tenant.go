package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/repository"
	"gorm.io/gorm"
)

// TenantAccessMiddleware validates that the user has access to the tenant
// Platform admins can access any tenant without membership
func TenantAccessMiddleware(memberRepo repository.MemberRepository, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, err := GetUserID(c)
		if err != nil {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get tenant ID from URL parameter
		tenantIDStr := c.Param("id")
		if tenantIDStr == "" {
			tenantIDStr = c.Param("tenant_id")
		}

		if tenantIDStr == "" {
			response.BadRequest(c, fmt.Errorf("tenant ID is required"))
			c.Abort()
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			response.BadRequest(c, fmt.Errorf("invalid tenant ID"))
			c.Abort()
			return
		}

		// Check if user is a platform admin - they can access any tenant
		var admin models.PlatformAdmin
		err = db.Where("user_id = ?", userID).First(&admin).Error
		if err == nil {
			// User is a platform admin - grant access without membership check
			c.Set("tenantID", tenantID)
			c.Set("isPlatformAdmin", true)
			c.Set("platformAdmin", &admin)
			c.Next()
			return
		}

		// Not a platform admin, check tenant membership
		member, err := memberRepo.GetByTenantAndUser(tenantID, userID)
		if err != nil || member == nil {
			response.Forbidden(c, "Access denied: You are not a member of this tenant")
			c.Abort()
			return
		}

		// Check if member is active
		if member.Status != "active" {
			response.Forbidden(c, "Access denied: Your membership is not active")
			c.Abort()
			return
		}

		// Store tenant ID and member in context for later use
		c.Set("tenantID", tenantID)
		c.Set("member", member)

		c.Next()
	}
}

// GetTenantID extracts the tenant ID from the Gin context
func GetTenantID(c *gin.Context) (uuid.UUID, error) {
	tenantID, exists := c.Get("tenantID")
	if !exists {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}

	tenantUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID is not a UUID")
	}

	return tenantUUID, nil
}
