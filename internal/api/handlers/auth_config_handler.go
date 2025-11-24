package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/config"
	"github.com/ysaakpr/rex/internal/pkg/response"
)

type AuthConfigHandler struct {
	config *config.Config
}

func NewAuthConfigHandler(cfg *config.Config) *AuthConfigHandler {
	return &AuthConfigHandler{
		config: cfg,
	}
}

type AuthConfigResponse struct {
	Providers ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	Google bool `json:"google"`
	// Future providers can be added here:
	// GitHub bool `json:"github"`
	// Microsoft bool `json:"microsoft"`
}

// GetAuthConfig returns the authentication configuration (which OAuth providers are enabled)
// This is a public endpoint that doesn't require authentication
func (h *AuthConfigHandler) GetAuthConfig(c *gin.Context) {
	authConfig := AuthConfigResponse{
		Providers: ProviderConfig{
			Google: h.config.IsGoogleOAuthEnabled(),
		},
	}

	response.OK(c, authConfig)
}

