package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	oauthService "github.com/julesChu12/fly/custos/internal/domain/service/oauth"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
)

type OAuthHandler struct {
	oauthService    *oauthService.Service
	userProfileRepo repository.UserProfileRepository
	tokenService    *token.TokenService
}

func NewOAuthHandler(oauthService *oauthService.Service, userProfileRepo repository.UserProfileRepository, tokenService *token.TokenService) *OAuthHandler {
	return &OAuthHandler{
		oauthService:    oauthService,
		userProfileRepo: userProfileRepo,
		tokenService:    tokenService,
	}
}

// GetOAuthURL generates OAuth authorization URL
// @Summary 获取OAuth授权URL
// @Description 生成指定OAuth提供商的授权URL（支持Google、GitHub）
// @Tags OAuth
// @Produce json
// @Param provider path string true "OAuth提供商（google/github）"
// @Param redirect_url query string true "回调地址"
// @Success 200 {object} object{auth_url=string,state=string}
// @Failure 400 {object} object{error=string}
// @Router /oauth/{provider}/login [get]
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
// @Summary 处理OAuth回调
// @Description 处理OAuth提供商的回调并完成登录
// @Tags OAuth
// @Produce json
// @Param provider path string true "OAuth提供商（google/github）"
// @Param code query string true "授权码"
// @Param state query string true "状态码"
// @Param redirect_url query string false "回调地址"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /oauth/{provider}/callback [get]
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

	// Create response
	userInfo := &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   string(user.Status),
		Role:     string(user.Role),
	}

	// Fetch profile data
	profile, err := h.userProfileRepo.GetByUserID(c.Request.Context(), user.ID)
	if err == nil {
		userInfo.Nickname = profile.Nickname
		userInfo.Avatar = profile.Avatar
	}

	response := dto.LoginResponse{
		User:             userInfo,
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
// @Summary 绑定OAuth账号
// @Description 将OAuth提供商账号绑定到当前登录用户
// @Tags OAuth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "OAuth提供商（google/github）"
// @Param request body dto.OAuthBindRequest true "绑定信息"
// @Success 200 {object} dto.OAuthBindResponse
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 409 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /oauth/{provider}/bind [post]
func (h *OAuthHandler) BindOAuthProvider(c *gin.Context) {
	provider := c.Param("provider")

	// Get authenticated user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID format",
		})
		return
	}

	var req dto.OAuthBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
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

	// Validate state parameter
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != req.State {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid state parameter",
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get OAuth config
	oauthConfig := h.oauthService.GetOAuthConfig(oauthProvider)
	if oauthConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "OAuth provider not configured",
		})
		return
	}

	// Set redirect URL
	oauthConfig.RedirectURL = req.RedirectURL

	// Exchange code for token
	token, err := oauthConfig.Exchange(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange code for token",
		})
		return
	}

	// Get user info from provider
	userInfo, err := h.oauthService.GetUserInfo(oauthProvider, token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user info from provider",
		})
		return
	}

	// Check if this OAuth account is already bound to another user
	existingOAuth, err := h.oauthService.GetOAuthByProviderUID(c.Request.Context(), string(oauthProvider), userInfo.ID)
	if err == nil && existingOAuth != nil && existingOAuth.UserID != uid {
		c.JSON(http.StatusConflict, gin.H{
			"error": "this OAuth account is already bound to another user",
		})
		return
	}

	// Bind OAuth provider to current user
	if err := h.oauthService.BindProvider(c.Request.Context(), uid, oauthProvider, userInfo.ID, token.AccessToken, token.RefreshToken, &token.Expiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to bind OAuth provider",
		})
		return
	}

	c.JSON(http.StatusOK, dto.OAuthBindResponse{
		Success: true,
		Message: fmt.Sprintf("%s account successfully bound", provider),
	})
}

// UnbindOAuthProvider unbinds OAuth provider from authenticated user
// @Summary 解绑OAuth账号
// @Description 将OAuth提供商账号从当前登录用户解绑
// @Tags OAuth
// @Produce json
// @Security BearerAuth
// @Param provider path string true "OAuth提供商（google/github）"
// @Success 200 {object} dto.OAuthUnbindResponse
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /oauth/{provider}/unbind [delete]
func (h *OAuthHandler) UnbindOAuthProvider(c *gin.Context) {
	provider := c.Param("provider")

	// Get authenticated user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID format",
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

	// Unbind OAuth provider
	if err := h.oauthService.UnbindProvider(c.Request.Context(), uid, oauthProvider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to unbind OAuth provider",
		})
		return
	}

	c.JSON(http.StatusOK, dto.OAuthUnbindResponse{
		Success: true,
		Message: fmt.Sprintf("%s account successfully unbound", provider),
	})
}

// GetUserOAuthBindings gets all OAuth bindings for authenticated user
// @Summary 获取OAuth绑定列表
// @Description 获取当前登录用户的所有OAuth账号绑定
// @Tags OAuth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.OAuthBindingsResponse
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /oauth/bindings [get]
func (h *OAuthHandler) GetUserOAuthBindings(c *gin.Context) {
	// Get authenticated user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID format",
		})
		return
	}

	// Get OAuth bindings
	bindings, err := h.oauthService.GetUserBindings(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get OAuth bindings",
		})
		return
	}

	// Convert to DTO
	bindingInfos := make([]dto.OAuthBindingInfo, 0, len(bindings))
	for _, binding := range bindings {
		bindingInfos = append(bindingInfos, dto.OAuthBindingInfo{
			ID:          binding.ID,
			Provider:    binding.Provider,
			ProviderUID: binding.ProviderUID,
			BoundAt:     binding.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, dto.OAuthBindingsResponse{
		Bindings: bindingInfos,
	})
}
