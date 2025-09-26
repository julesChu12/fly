package oauth

import (
	"context"
	"fmt"

	"github.com/julesChu12/custos/internal/domain/entity"
	"github.com/julesChu12/custos/internal/domain/repository"
	"github.com/julesChu12/custos/internal/domain/service/token"
	"github.com/julesChu12/custos/pkg/errors"
)

type OAuthUseCase struct {
	userRepo     repository.UserRepository
	tokenService *token.TokenService
}

func NewOAuthUseCase(userRepo repository.UserRepository, tokenService *token.TokenService) *OAuthUseCase {
	return &OAuthUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// OAuthProvider represents an OAuth provider
type OAuthProvider string

const (
	ProviderGoogle    OAuthProvider = "google"
	ProviderGitHub    OAuthProvider = "github"
	ProviderMicrosoft OAuthProvider = "microsoft"
)

// OAuthUserInfo contains user information from OAuth provider
type OAuthUserInfo struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

// AuthorizeWithOAuth handles OAuth authorization flow
func (uc *OAuthUseCase) AuthorizeWithOAuth(ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo) (*token.TokenPair, *entity.User, error) {
	// Check if user exists by email
	user, err := uc.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		// User doesn't exist, create new user
		user = entity.NewUser(userInfo.Name, userInfo.Email, "")
		user.SetOAuthProvider(string(provider), userInfo.ProviderID)

		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("failed to create OAuth user: %w", err)
		}
	} else {
		// Update existing user with OAuth info
		user.SetOAuthProvider(string(provider), userInfo.ProviderID)
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("failed to update user OAuth info: %w", err)
		}
	}

	// Generate token for the user
	tokenPair, err := uc.tokenService.GenerateAccessToken(uc.tokenService.GenerateSessionID(), user.ID, user.Username, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate OAuth token: %w", err)
	}

	return tokenPair, user, nil
}

// GetOAuthURL generates OAuth authorization URL
func (uc *OAuthUseCase) GetOAuthURL(ctx context.Context, provider OAuthProvider, state string) (string, error) {
	// This would typically generate the OAuth authorization URL
	// For now, return a placeholder
	switch provider {
	case ProviderGoogle:
		return fmt.Sprintf("https://accounts.google.com/oauth/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&scope=openid email profile&state=%s", state), nil
	case ProviderGitHub:
		return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&scope=user:email&state=%s", state), nil
	case ProviderMicrosoft:
		return fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&scope=openid email profile&state=%s", state), nil
	default:
		return "", errors.NewInvalidProviderError(string(provider))
	}
}

// RevokeOAuthToken revokes OAuth access token
func (uc *OAuthUseCase) RevokeOAuthToken(ctx context.Context, provider OAuthProvider, token string) error {
	// This would typically call the OAuth provider's token revocation endpoint
	// For now, just return success
	return nil
}
