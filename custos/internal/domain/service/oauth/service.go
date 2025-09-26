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
	"net/url"
	"strconv"
	"strings"
	"time"

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
}

func NewService(cfg *config.Config, userRepo repository.UserRepository, userOAuthRepo repository.UserOAuthRepository) *Service {
	return &Service{
		cfg:           cfg,
		userRepo:      userRepo,
		userOAuthRepo: userOAuthRepo,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// GenerateAuthURL generates OAuth authorization URL with state
func (s *Service) GenerateAuthURL(ctx context.Context, provider Provider, redirectURL string) (string, string, error) {
	var providerConfig config.OAuthProvider
	switch provider {
	case Google:
		providerConfig = s.cfg.OAuth.Google
	case GitHub:
		providerConfig = s.cfg.OAuth.GitHub
	default:
		return "", "", errors.NewInvalidProviderError(string(provider))
	}

	// Generate state parameter
	state := s.generateState()

	// Build auth URL
	params := url.Values{}
	params.Set("client_id", providerConfig.ClientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("scope", strings.Join(providerConfig.Scopes, " "))
	params.Set("state", state)
	params.Set("response_type", "code")

	if provider == Google {
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	}

	authURL := providerConfig.AuthURL + "?" + params.Encode()
	return authURL, state, nil
}

// HandleCallback handles OAuth callback and creates/updates user
func (s *Service) HandleCallback(ctx context.Context, provider Provider, code, state, redirectURL string) (*entity.User, *entity.UserOAuth, error) {
	// Validate state (in production, you should store and validate state properly)
	if !s.validateState(state) {
		return nil, nil, fmt.Errorf("invalid state parameter")
	}

	var providerConfig config.OAuthProvider
	switch provider {
	case Google:
		providerConfig = s.cfg.OAuth.Google
	case GitHub:
		providerConfig = s.cfg.OAuth.GitHub
	default:
		return nil, nil, errors.NewInvalidProviderError(string(provider))
	}

	// Exchange code for token
	token, err := s.exchangeCodeForToken(providerConfig, code, redirectURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfo(providerConfig, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if OAuth binding exists
	userOAuth, err := s.userOAuthRepo.GetByProviderUID(ctx, string(provider), userInfo.ID)
	if err != nil {
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
		userOAuth.UpdateTokens(token.AccessToken, token.RefreshToken, token.ExpiresAt)
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
		userOAuth.UpdateTokens(token.AccessToken, token.RefreshToken, token.ExpiresAt)

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

type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int        `json:"expires_in"`
	ExpiresAt    *time.Time `json:"-"`
}

func (s *Service) exchangeCodeForToken(providerConfig config.OAuthProvider, code, redirectURL string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", providerConfig.ClientID)
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", providerConfig.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	// Calculate expiration time
	if token.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		token.ExpiresAt = &expiresAt
	}

	return &token, nil
}

func (s *Service) getUserInfo(providerConfig config.OAuthProvider, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", providerConfig.UserInfoURL, nil)
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

	return &userInfo, nil
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