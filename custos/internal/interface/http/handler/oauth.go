package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	oauthService "github.com/julesChu12/fly/custos/internal/domain/service/oauth"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
)

type OAuthHandler struct {
	oauthService *oauthService.Service
	tokenService *token.TokenService
}

func NewOAuthHandler(oauthService *oauthService.Service, tokenService *token.TokenService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		tokenService: tokenService,
	}
}

// GetOAuthURL generates OAuth authorization URL
// GET /api/v1/oauth/{provider}/login
func (h *OAuthHandler) GetOAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	redirectURL := c.Query("redirect_url")

	if redirectURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "redirect_url parameter is required",
		})
		return
	}

	var oauthProvider oauthService.Provider
	switch strings.ToLower(provider) {
	case "google":
		oauthProvider = oauthService.Google
	case "github":
		oauthProvider = oauthService.GitHub
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported OAuth provider",
		})
		return
	}

	authURL, state, err := h.oauthService.GenerateAuthURL(c.Request.Context(), oauthProvider, redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate OAuth URL",
		})
		return
	}

	// Store state in cookie for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleOAuthCallback handles OAuth callback from provider
// GET /api/v1/oauth/{provider}/callback
func (h *OAuthHandler) HandleOAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	redirectURL := c.Query("redirect_url")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "authorization code is required",
		})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "state parameter is required",
		})
		return
	}

	// Validate state from cookie
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid state parameter",
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	var oauthProvider oauthService.Provider
	switch strings.ToLower(provider) {
	case "google":
		oauthProvider = oauthService.Google
	case "github":
		oauthProvider = oauthService.GitHub
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported OAuth provider",
		})
		return
	}

	if redirectURL == "" {
		// Use default redirect URL or construct from request
		redirectURL = c.Request.Header.Get("Referer")
		if redirectURL == "" {
			redirectURL = "http://localhost:8080/api/v1/oauth/" + provider + "/callback"
		}
	}

	user, _, err := h.oauthService.HandleCallback(c.Request.Context(), oauthProvider, code, state, redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "OAuth callback processing failed",
		})
		return
	}

	// Generate internal JWT tokens
	tokenPair, err := h.tokenService.GenerateAccessToken(
		h.tokenService.GenerateSessionID(),
		user.ID,
		user.Username,
		user.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate access token",
		})
		return
	}

	response := dto.LoginResponse{
		User: &dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   string(user.Status),
			Role:     string(user.Role),
		},
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		ExpiresIn:        900,    // 15 minutes in seconds
		RefreshExpiresIn: 604800, // 7 days in seconds
		TokenType:        "Bearer",
		SessionID:        tokenPair.SessionID,
	}

	c.JSON(http.StatusOK, response)
}

// BindOAuthProvider binds OAuth provider to existing authenticated user
// POST /api/v1/oauth/{provider}/bind
func (h *OAuthHandler) BindOAuthProvider(c *gin.Context) {
	// This would require authentication middleware to get current user
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "OAuth provider binding not implemented yet",
	})
}

// UnbindOAuthProvider unbinds OAuth provider from authenticated user
// DELETE /api/v1/oauth/{provider}/unbind
func (h *OAuthHandler) UnbindOAuthProvider(c *gin.Context) {
	// This would require authentication middleware to get current user
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "OAuth provider unbinding not implemented yet",
	})
}

// GetUserOAuthBindings gets all OAuth bindings for authenticated user
// GET /api/v1/oauth/bindings
func (h *OAuthHandler) GetUserOAuthBindings(c *gin.Context) {
	// This would require authentication middleware to get current user
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "OAuth bindings listing not implemented yet",
	})
}
