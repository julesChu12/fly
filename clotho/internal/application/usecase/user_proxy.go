package usecase

import (
	"context"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// UserProxyUseCase handles user-related operations by orchestrating calls to Custos service
type UserProxyUseCase struct {
	custosClient *client.CustosClient
	timeout      time.Duration
}

// NewUserProxyUseCase creates a new UserProxyUseCase instance
func NewUserProxyUseCase(custosClient *client.CustosClient, timeout time.Duration) *UserProxyUseCase {
	return &UserProxyUseCase{
		custosClient: custosClient,
		timeout:      timeout,
	}
}

// GetUserByID retrieves user information by user ID from Custos service
func (u *UserProxyUseCase) GetUserByID(userID int64) (*client.UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	userInfo, err := u.custosClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// ValidateUserToken validates a user token with Custos service
func (u *UserProxyUseCase) ValidateUserToken(token string) (*client.UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	userInfo, err := u.custosClient.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// GetCurrentUserProfile retrieves the current user's profile information
// This is an example of how Clotho orchestrates multiple calls if needed
func (u *UserProxyUseCase) GetCurrentUserProfile(userID int64) (*UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	// Get user basic info from Custos
	userInfo, err := u.custosClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// TODO: In the future, this could aggregate data from multiple services
	// For example, get user preferences from another service, order history, etc.

	profile := &UserProfile{
		User:        userInfo,
		Preferences: nil, // Could come from another service
		Statistics:  nil, // Could come from analytics service
	}

	return profile, nil
}

// UserProfile represents an aggregated user profile with data from multiple services
type UserProfile struct {
	User        *client.UserInfo  `json:"user"`
	Preferences map[string]string `json:"preferences,omitempty"`
	Statistics  map[string]int64  `json:"statistics,omitempty"`
}
