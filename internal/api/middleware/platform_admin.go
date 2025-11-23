package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/pkg/response"
	"gorm.io/gorm"
)

// PlatformAdminMiddleware checks if the user is a platform admin
func PlatformAdminMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Check if user is a platform admin
		var admin models.PlatformAdmin
		if err := db.Where("user_id = ?", userID).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Forbidden(c, "Platform admin access required")
				c.Abort()
				return
			}
			response.InternalServerError(c, err)
			c.Abort()
			return
		}

		// Store admin info in context
		c.Set("platformAdmin", &admin)
		c.Next()
	}
}

// GetPlatformAdmin retrieves platform admin from context
func GetPlatformAdmin(c *gin.Context) (*models.PlatformAdmin, error) {
	admin, exists := c.Get("platformAdmin")
	if !exists {
		return nil, nil
	}
	return admin.(*models.PlatformAdmin), nil
}
