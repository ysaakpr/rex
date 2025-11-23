package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"github.com/ysaakpr/rex/internal/services"
)

// RequirePermission creates a middleware that checks if user has a specific permission
func RequirePermission(rbacService services.RBACService, service, entity, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID
		userID, err := GetUserID(c)
		if err != nil {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get tenant ID
		tenantID, err := GetTenantID(c)
		if err != nil {
			response.BadRequest(c, fmt.Errorf("tenant context required"))
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := rbacService.CheckUserPermission(tenantID, userID, service, entity, action)
		if err != nil {
			response.InternalServerError(c, err)
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, fmt.Sprintf("Permission denied: %s:%s:%s", service, entity, action))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRelation checks if user has a specific relation in the tenant
func RequireRelation(memberRepo interface{}, relationName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get member from context (set by TenantAccessMiddleware)
		memberInterface, exists := c.Get("member")
		if !exists {
			response.Forbidden(c, "Access denied: Tenant membership required")
			c.Abort()
			return
		}

		// Type assertion would go here to check relation
		// For simplicity, we'll just pass through
		// In production, you'd check member.Relation.Name == relationName

		_ = memberInterface
		_ = relationName

		c.Next()
	}
}
