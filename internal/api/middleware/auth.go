package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
	"github.com/ysaakpr/rex/internal/pkg/response"
)

// AuthMiddleware verifies SuperTokens session
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log incoming request for debugging
		fmt.Printf("[DEBUG] AuthMiddleware: Path=%s, Host=%s, Cookies=%d\n", c.Request.URL.Path, c.Request.Host, len(c.Request.Cookies()))

		// Track if verification succeeded
		verificationSucceeded := false

		session.VerifySession(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[DEBUG] Inside VerifySession handler - session verified!\n")
			verificationSucceeded = true
			c.Request = c.Request.WithContext(r.Context())
		})).ServeHTTP(c.Writer, c.Request)

		if !verificationSucceeded {
			// Session verification failed - SuperTokens already wrote the error response
			fmt.Printf("[DEBUG] Session verification failed - SuperTokens returned error\n")
			c.Abort()
			return
		}

		// Session verified successfully, extract user ID
		sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
		if sessionContainer == nil {
			fmt.Printf("[DEBUG] Session container is nil after verification\n")
			response.Unauthorized(c, "Session not found")
			c.Abort()
			return
		}

		userID := sessionContainer.GetUserID()
		fmt.Printf("[DEBUG] Session verified successfully for user: %s\n", userID)
		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware checks for session but doesn't require it
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
		if sessionContainer != nil {
			userID := sessionContainer.GetUserID()
			c.Set("userID", userID)
		}
		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user ID is not a string")
	}

	return userIDStr, nil
}

// GetUserEmail fetches the user's email from SuperTokens
func GetUserEmail(c *gin.Context) (string, error) {
	userID, err := GetUserID(c)
	if err != nil {
		return "", err
	}

	// Get user info from SuperTokens
	userInfo, err := emailpassword.GetUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	if userInfo == nil {
		return "", fmt.Errorf("user not found")
	}

	return userInfo.Email, nil
}

// GetSession returns the SuperTokens session from context
func GetSession(c *gin.Context) (sessmodels.SessionContainer, error) {
	sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
	if sessionContainer == nil {
		return nil, fmt.Errorf("session not found")
	}
	return sessionContainer, nil
}

// SuperTokensMiddleware wraps SuperTokens middleware for Gin
func SuperTokensMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		supertokens.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = c.Request.WithContext(r.Context())
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)

		// If SuperTokens middleware handled the request, we're done
		if c.Writer.Written() {
			c.Abort()
			return
		}
	}
}
