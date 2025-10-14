package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth_GetOAuthURL_Google(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Get OAuth URL for Google
	w := ts.MakeRequest(http.MethodGet, "/api/v1/oauth/google/login", nil, nil)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get OAuth URL: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	authURL, ok := response["url"].(string)
	assert.True(t, ok)
	assert.Contains(t, authURL, "accounts.google.com")
	assert.Contains(t, authURL, "state=")
}

func TestOAuth_GetOAuthURL_GitHub(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Get OAuth URL for GitHub
	w := ts.MakeRequest(http.MethodGet, "/api/v1/oauth/github/login", nil, nil)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get OAuth URL: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	authURL, ok := response["url"].(string)
	assert.True(t, ok)
	assert.Contains(t, authURL, "github.com")
	assert.Contains(t, authURL, "state=")
}

func TestOAuth_GetOAuthURL_InvalidProvider(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try invalid provider
	w := ts.MakeRequest(http.MethodGet, "/api/v1/oauth/invalid_provider/login", nil, nil)

	AssertErrorResponse(t, w, http.StatusBadRequest, "provider")
}

func TestOAuth_BindProvider_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_bind@test.com"
	password := "TestPass123!"
	registerResp := ts.RegisterTestUser(t, email, password)
	accessToken := registerResp["access_token"].(string)
	userID := uint(registerResp["user_id"].(float64))

	// Simulate OAuth binding by directly creating the binding
	// (In real integration tests with OAuth providers, you'd go through the full flow)
	expiresAt := time.Now().Add(1 * time.Hour)
	oauthBinding := &entity.UserOAuth{
		UserID:       userID,
		Provider:     "google",
		ProviderUID:  "google_test_uid_123",
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    &expiresAt,
	}

	err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding)
	require.NoError(t, err)

	// Verify binding was created by getting bindings
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get bindings: %s", w.Body.String())

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	bindings := response["bindings"].([]interface{})
	assert.Len(t, bindings, 1)

	binding := bindings[0].(map[string]interface{})
	assert.Equal(t, "google", binding["provider"])
	assert.Equal(t, "oauth_bind@gmail.com", binding["email"])
}

func TestOAuth_UnbindProvider_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_unbind@test.com"
	password := "TestPass123!"
	registerResp := ts.RegisterTestUser(t, email, password)
	accessToken := registerResp["access_token"].(string)
	userID := uint(registerResp["user_id"].(float64))

	// Create OAuth binding
	expiresAt := time.Now().Add(1 * time.Hour)
	oauthBinding := &entity.UserOAuth{
		UserID:       userID,
		Provider:     "github",
		ProviderUID:  "github_test_uid_456",
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    &expiresAt,
	}

	err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding)
	require.NoError(t, err)

	// Verify binding exists
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var beforeResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &beforeResponse)
	require.NoError(t, err)
	assert.Len(t, beforeResponse["bindings"].([]interface{}), 1)

	// Unbind provider
	w = ts.MakeAuthenticatedRequest(http.MethodDelete, "/api/v1/oauth/github/unbind", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to unbind: %s", w.Body.String())

	// Verify binding was removed
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var afterResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &afterResponse)
	require.NoError(t, err)
	assert.Len(t, afterResponse["bindings"].([]interface{}), 0)
}

func TestOAuth_UnbindProvider_NotBound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_unbind_notbound@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Try to unbind a provider that was never bound
	w := ts.MakeAuthenticatedRequest(http.MethodDelete, "/api/v1/oauth/google/unbind", nil, accessToken)

	AssertErrorResponse(t, w, http.StatusNotFound, "not found")
}

func TestOAuth_GetBindings_Empty(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_no_bindings@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Get bindings (should be empty)
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get bindings: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	bindings := response["bindings"].([]interface{})
	assert.Len(t, bindings, 0)
}

func TestOAuth_GetBindings_Multiple(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_multiple@test.com"
	password := "TestPass123!"
	registerResp := ts.RegisterTestUser(t, email, password)
	accessToken := registerResp["access_token"].(string)
	userID := uint(registerResp["user_id"].(float64))

	// Create multiple OAuth bindings
	providers := []struct {
		provider    string
		providerUID string
		email       string
	}{
		{"google", "google_uid_123", "user@gmail.com"},
		{"github", "github_uid_456", "user@github.com"},
	}

	for _, p := range providers {
		expiresAt := time.Now().Add(1 * time.Hour)
		oauthBinding := &entity.UserOAuth{
			UserID:       userID,
			Provider:     p.provider,
			ProviderUID:  p.providerUID,
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    &expiresAt,
		}

		err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding)
		require.NoError(t, err)
	}

	// Get all bindings
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get bindings: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	bindings := response["bindings"].([]interface{})
	assert.Len(t, bindings, 2)

	// Verify both providers are present
	providerSet := make(map[string]bool)
	for _, b := range bindings {
		binding := b.(map[string]interface{})
		providerSet[binding["provider"].(string)] = true
	}

	assert.True(t, providerSet["google"])
	assert.True(t, providerSet["github"])
}

func TestOAuth_Bindings_Unauthenticated(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to get bindings without authentication
	w := ts.MakeRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "missing or invalid")
}

func TestOAuth_BindProvider_Unauthenticated(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to bind without authentication
	reqBody := map[string]string{
		"code":  "test_code",
		"state": "test_state",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/oauth/google/bind", reqBody, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "missing or invalid")
}

func TestOAuth_MultipleUsersIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create two users
	user1Resp := ts.RegisterTestUser(t, "user1@test.com", "Pass123!")
	user1Token := user1Resp["access_token"].(string)
	user1ID := uint(user1Resp["user_id"].(float64))

	user2Resp := ts.RegisterTestUser(t, "user2@test.com", "Pass123!")
	user2Token := user2Resp["access_token"].(string)
	user2ID := uint(user2Resp["user_id"].(float64))

	// Create OAuth bindings for both users with same provider
	for i, userID := range []uint{user1ID, user2ID} {
		expiresAt := time.Now().Add(1 * time.Hour)
		oauthBinding := &entity.UserOAuth{
			UserID:       userID,
			Provider:     "google",
			ProviderUID:  fmt.Sprintf("google_uid_%d", i+1),
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    &expiresAt,
		}

		err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding)
		require.NoError(t, err)
	}

	// Get bindings for user1
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, user1Token)
	require.Equal(t, http.StatusOK, w.Code)

	var user1Bindings map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user1Bindings)
	require.NoError(t, err)

	user1BindingsList := user1Bindings["bindings"].([]interface{})
	assert.Len(t, user1BindingsList, 1)

	user1Binding := user1BindingsList[0].(map[string]interface{})
	assert.Equal(t, "user1@gmail.com", user1Binding["email"])

	// Get bindings for user2
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, user2Token)
	require.Equal(t, http.StatusOK, w.Code)

	var user2Bindings map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &user2Bindings)
	require.NoError(t, err)

	user2BindingsList := user2Bindings["bindings"].([]interface{})
	assert.Len(t, user2BindingsList, 1)

	user2Binding := user2BindingsList[0].(map[string]interface{})
	assert.Equal(t, "user2@gmail.com", user2Binding["email"])

	// Ensure bindings are isolated
	assert.NotEqual(t, user1Binding["email"], user2Binding["email"])
}

func TestOAuth_DuplicateBinding_SameProvider(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_duplicate@test.com"
	password := "TestPass123!"
	registerResp := ts.RegisterTestUser(t, email, password)
	userID := uint(registerResp["user_id"].(float64))

	// Create first binding
	expiresAt1 := time.Now().Add(1 * time.Hour)
	oauthBinding1 := &entity.UserOAuth{
		UserID:       userID,
		Provider:     "google",
		ProviderUID:  "google_uid_123",
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    &expiresAt1,
	}

	err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding1)
	require.NoError(t, err)

	// Try to create duplicate binding with same provider
	expiresAt2 := time.Now().Add(1 * time.Hour)
	oauthBinding2 := &entity.UserOAuth{
		UserID:       userID,
		Provider:     "google",
		ProviderUID:  "google_uid_456", // Different UID
		AccessToken:  "mock_access_token_2",
		RefreshToken: "mock_refresh_token_2",
		ExpiresAt:    &expiresAt2,
	}

	// This should fail or update the existing binding depending on your business logic
	err = ts.UserOAuthRepo.Create(context.Background(), oauthBinding2)

	// Document the behavior - adjust based on your requirements
	if err != nil {
		t.Log("System prevents duplicate bindings for same provider (constraint enforced)")
	} else {
		t.Log("System allows overwriting existing binding (no constraint)")
	}
}

func TestOAuth_TokenExpiration(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "oauth_expired@test.com"
	password := "TestPass123!"
	registerResp := ts.RegisterTestUser(t, email, password)
	accessToken := registerResp["access_token"].(string)
	userID := uint(registerResp["user_id"].(float64))

	// Create binding with expired OAuth token
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	oauthBinding := &entity.UserOAuth{
		UserID:       userID,
		Provider:     "google",
		ProviderUID:  "google_uid_expired",
		AccessToken:  "expired_access_token",
		RefreshToken: "expired_refresh_token",
		ExpiresAt:    &expiresAt,
	}

	err := ts.UserOAuthRepo.Create(context.Background(), oauthBinding)
	require.NoError(t, err)

	// Get bindings - should still return the binding
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/oauth/bindings", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	bindings := response["bindings"].([]interface{})
	assert.Len(t, bindings, 1)

	// Note: The binding is returned even though the OAuth token is expired
	// Token refresh would happen when the user tries to use the OAuth provider again
}
