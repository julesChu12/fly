package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/julesChu12/fly/custos/internal/config"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/pkg/errors"
)

type Provider string

const (
	Google Provider = "google"
	GitHub Provider = "github"
)

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Verified bool   `json:"email_verified"`
}

type Service struct {
	cfg           *config.Config
	userRepo      repository.UserRepository
	userOAuthRepo repository.UserOAuthRepository
	httpClient    *http.Client
	oauthConfigs  map[Provider]*oauth2.Config
}

func NewService(cfg *config.Config, userRepo repository.UserRepository, userOAuthRepo repository.UserOAuthRepository) *Service {
	s := &Service{
		cfg:           cfg,
		userRepo:      userRepo,
		userOAuthRepo: userOAuthRepo,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		oauthConfigs:  make(map[Provider]*oauth2.Config),
	}

	// Initialize OAuth configs
	s.initOAuthConfigs()
	return s
}

// initOAuthConfigs initializes OAuth2 configurations for different providers
func (s *Service) initOAuthConfigs() {
	// Google OAuth config
	if s.cfg.OAuth.Google.ClientID != "" {
		s.oauthConfigs[Google] = &oauth2.Config{
			ClientID:     s.cfg.OAuth.Google.ClientID,
			ClientSecret: s.cfg.OAuth.Google.ClientSecret,
			Scopes:       s.cfg.OAuth.Google.Scopes,
			Endpoint:     google.Endpoint,
		}
	}

	// GitHub OAuth config
	if s.cfg.OAuth.GitHub.ClientID != "" {
		s.oauthConfigs[GitHub] = &oauth2.Config{
			ClientID:     s.cfg.OAuth.GitHub.ClientID,
			ClientSecret: s.cfg.OAuth.GitHub.ClientSecret,
			Scopes:       s.cfg.OAuth.GitHub.Scopes,
			Endpoint:     github.Endpoint,
		}
	}
}

// GenerateAuthURL generates OAuth authorization URL with state
func (s *Service) GenerateAuthURL(ctx context.Context, provider Provider, redirectURL string) (string, string, error) {
	oauthConfig, exists := s.oauthConfigs[provider]
	if !exists {
		return "", "", errors.NewInvalidProviderError(string(provider))
	}

	// Set redirect URL
	oauthConfig.RedirectURL = redirectURL

	// Generate state parameter
	state := s.generateState()

	// Generate authorization URL
	var authURL string
	if provider == Google {
		authURL = oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	} else {
		authURL = oauthConfig.AuthCodeURL(state)
	}

	return authURL, state, nil
}

// HandleCallback handles OAuth callback and creates/updates user
func (s *Service) HandleCallback(ctx context.Context, provider Provider, code, state, redirectURL string) (*entity.User, *entity.UserOAuth, error) {
	// Validate state (in production, you should store and validate state properly)
	if !s.validateState(state) {
		return nil, nil, fmt.Errorf("invalid state parameter")
	}

	oauthConfig, exists := s.oauthConfigs[provider]
	if !exists {
		return nil, nil, errors.NewInvalidProviderError(string(provider))
	}

	// Set redirect URL
	oauthConfig.RedirectURL = redirectURL

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfo(provider, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if OAuth binding exists
	userOAuth, err := s.userOAuthRepo.GetByProviderUID(ctx, string(provider), userInfo.ID)
	if err != nil && err != repository.ErrUserOAuthNotFound {
		return nil, nil, fmt.Errorf("failed to check existing OAuth binding: %w", err)
	}

	var user *entity.User

	if userOAuth != nil {
		// Existing OAuth binding - get associated user
		user, err = s.userRepo.GetByID(ctx, userOAuth.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Update OAuth tokens
		var expiresAt *time.Time
		if token.Expiry != (time.Time{}) {
			expiresAt = &token.Expiry
		}
		userOAuth.UpdateTokens(token.AccessToken, token.RefreshToken, expiresAt)
		if err := s.userOAuthRepo.Update(ctx, userOAuth); err != nil {
			return nil, nil, fmt.Errorf("failed to update OAuth binding: %w", err)
		}
	} else {
		// No existing OAuth binding - check if user exists by email
		user, err = s.userRepo.GetByEmail(ctx, userInfo.Email)
		if err != nil && err != repository.ErrUserNotFound {
			return nil, nil, fmt.Errorf("failed to check user by email: %w", err)
		}

		if user == nil {
			// Create new user
			user = entity.NewUser("", userInfo.Email, "")
			user.Nickname = userInfo.Name
			user.Avatar = userInfo.Picture

			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, nil, fmt.Errorf("failed to create user: %w", err)
			}
		}

		// Create OAuth binding
		userOAuth = entity.NewUserOAuth(user.ID, string(provider), userInfo.ID)
		var expiresAt *time.Time
		if token.Expiry != (time.Time{}) {
			expiresAt = &token.Expiry
		}
		userOAuth.UpdateTokens(token.AccessToken, token.RefreshToken, expiresAt)

		if err := s.userOAuthRepo.Create(ctx, userOAuth); err != nil {
			return nil, nil, fmt.Errorf("failed to create OAuth binding: %w", err)
		}
	}

	return user, userOAuth, nil
}

// UnbindProvider unbinds OAuth provider from user
func (s *Service) UnbindProvider(ctx context.Context, userID uint, provider Provider) error {
	return s.userOAuthRepo.UnbindProvider(ctx, userID, string(provider))
}

// GetUserBindings gets all OAuth bindings for a user
func (s *Service) GetUserBindings(ctx context.Context, userID uint) ([]*entity.UserOAuth, error) {
	return s.userOAuthRepo.GetByUserID(ctx, userID)
}

func (s *Service) getUserInfo(provider Provider, accessToken string) (*UserInfo, error) {
	var userInfoURL string

	switch provider {
	case Google:
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case GitHub:
		userInfoURL = "https://api.github.com/user"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// Normalize response for different providers
	if provider == GitHub {
		// GitHub uses "login" for username and doesn't have email_verified
		if userInfo.Name == "" {
			userInfo.Name = userInfo.ID // GitHub login name
		}
		userInfo.Verified = true // Assume GitHub emails are verified

		// GitHub might not include email in the response, need separate call
		if userInfo.Email == "" {
			email, err := s.getGitHubUserEmail(accessToken)
			if err == nil {
				userInfo.Email = email
			}
		}
	}

	return &userInfo, nil
}

// getGitHubUserEmail gets the primary email from GitHub API
func (s *Service) getGitHubUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("email request failed with status: %d", resp.StatusCode)
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

func (s *Service) generateState() string {
	// Generate random bytes
	b := make([]byte, 32)
	rand.Read(b)

	// Create HMAC with timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	h := hmac.New(sha256.New, []byte(s.cfg.OAuth.StateKey))
	h.Write([]byte(timestamp))
	h.Write(b)

	// Combine timestamp and MAC
	state := timestamp + ":" + base64.URLEncoding.EncodeToString(h.Sum(nil))
	return base64.URLEncoding.EncodeToString([]byte(state))
}

func (s *Service) validateState(state string) bool {
	// Decode state
	decoded, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}

	// Check if state is expired
	if time.Now().Unix()-timestamp > int64(s.cfg.OAuth.StateTTL) {
		return false
	}

	// Validate HMAC
	expectedMAC, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(s.cfg.OAuth.StateKey))
	h.Write([]byte(parts[0]))

	return hmac.Equal(expectedMAC, h.Sum(nil))
}