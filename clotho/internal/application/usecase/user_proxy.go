package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// UserProxyUseCase handles user-related operations by orchestrating calls to Custos service
type UserProxyUseCase struct {
	custosClient  *client.CustosClient
	timeout       time.Duration
	clientFactory func() (*client.CustosClient, error)
	clientMu      sync.Mutex
}

// NewUserProxyUseCase creates a new UserProxyUseCase instance
func NewUserProxyUseCase(custosClient *client.CustosClient, timeout time.Duration) *UserProxyUseCase {
	return &UserProxyUseCase{
		custosClient: custosClient,
		timeout:      timeout,
	}
}

// SetCustosClientFactory configures lazy initialization for the Custos client
func (u *UserProxyUseCase) SetCustosClientFactory(factory func() (*client.CustosClient, error)) {
	u.clientMu.Lock()
	defer u.clientMu.Unlock()
	u.clientFactory = factory
}

func (u *UserProxyUseCase) getCustosClient() (*client.CustosClient, error) {
	if u.custosClient != nil {
		return u.custosClient, nil
	}

	u.clientMu.Lock()
	defer u.clientMu.Unlock()

	if u.custosClient != nil {
		return u.custosClient, nil
	}

	if u.clientFactory == nil {
		return nil, errors.New("custos client factory not configured")
	}

	clientInstance, err := u.clientFactory()
	if err != nil {
		return nil, err
	}

	u.custosClient = clientInstance
	return u.custosClient, nil
}

// GetUserByID retrieves user information by user ID from Custos service
func (u *UserProxyUseCase) GetUserByID(userID int64) (*client.UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	custosClient, err := u.getCustosClient()
	if err != nil {
		return nil, err
	}

	userInfo, err := custosClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// ValidateUserToken validates a user token with Custos service
func (u *UserProxyUseCase) ValidateUserToken(token string) (*client.UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	custosClient, err := u.getCustosClient()
	if err != nil {
		return nil, err
	}

	userInfo, err := custosClient.ValidateToken(ctx, token)
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
	custosClient, err := u.getCustosClient()
	if err != nil {
		return nil, err
	}

	userInfo, err := custosClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user preferences from Custos
	preferences, err := custosClient.GetUserPreferences(ctx, userID)
	if err != nil {
		// If preferences service is not available, use default preferences
		preferences = map[string]string{
			"language":      "en",
			"timezone":      "UTC",
			"theme":         "light",
			"notifications": "enabled",
		}
	}

	// Get user statistics from Custos
	statistics, err := custosClient.GetUserStatistics(ctx, userID)
	if err != nil {
		// If statistics service is not available, use empty statistics
		statistics = map[string]int64{}
	}

	profile := &UserProfile{
		User:        userInfo,
		Preferences: preferences,
		Statistics:  statistics,
	}

	return profile, nil
}

// UpdateUserProfile updates user profile information through Custos service
func (u *UserProxyUseCase) UpdateUserProfile(userID int64, updates map[string]interface{}) (*UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	custosClient, err := u.getCustosClient()
	if err != nil {
		return nil, err
	}

	// Get current user info first
	userInfo, err := custosClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates to user info
	updateMask := []string{}
	if username, ok := updates["username"].(string); ok && username != "" {
		userInfo.Username = username
		updateMask = append(updateMask, "username")
	}
	if email, ok := updates["email"].(string); ok && email != "" {
		userInfo.Email = email
		updateMask = append(updateMask, "email")
	}
	if firstName, ok := updates["first_name"].(string); ok {
		userInfo.FirstName = firstName
		updateMask = append(updateMask, "first_name")
	}
	if lastName, ok := updates["last_name"].(string); ok {
		userInfo.LastName = lastName
		updateMask = append(updateMask, "last_name")
	}
	if avatar, ok := updates["avatar"].(string); ok {
		userInfo.Avatar = avatar
		updateMask = append(updateMask, "avatar")
	}
	if bio, ok := updates["bio"].(string); ok {
		userInfo.Bio = bio
		updateMask = append(updateMask, "bio")
	}
	if phone, ok := updates["phone"].(string); ok {
		userInfo.Phone = phone
		updateMask = append(updateMask, "phone")
	}
	if location, ok := updates["location"].(string); ok {
		userInfo.Location = location
		updateMask = append(updateMask, "location")
	}
	if website, ok := updates["website"].(string); ok {
		userInfo.Website = website
		updateMask = append(updateMask, "website")
	}

	// Update user profile through Custos service
	if len(updateMask) > 0 {
		_, err = custosClient.UpdateUser(ctx, userID, userInfo, updateMask)
		if err != nil {
			return nil, fmt.Errorf("failed to update user profile: %w", err)
		}
	}

	// Return updated profile
	return u.GetCurrentUserProfile(userID)
}

// UpdateUserPreferences updates user preferences through Custos service
func (u *UserProxyUseCase) UpdateUserPreferences(userID int64, preferences map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	custosClient, err := u.getCustosClient()
	if err != nil {
		return err
	}

	_, err = custosClient.UpdateUserPreferences(ctx, userID, preferences)
	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// GetUserStatistics retrieves user statistics from Custos analytics service
func (u *UserProxyUseCase) GetUserStatistics(userID int64) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	custosClient, err := u.getCustosClient()
	if err != nil {
		return nil, err
	}

	statistics, err := custosClient.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	return statistics, nil
}

// UserProfile represents an aggregated user profile with data from multiple services
type UserProfile struct {
	User        *client.UserInfo  `json:"user"`
	Preferences map[string]string `json:"preferences,omitempty"`
	Statistics  map[string]int64  `json:"statistics,omitempty"`
}
