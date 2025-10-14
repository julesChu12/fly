package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	reqBody := map[string]string{
		"email":    "newuser@test.com",
		"password": "SecurePass123!",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/register", reqBody, nil)

	require.Equal(t, http.StatusOK, w.Code, "Failed to register: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["user_id"])
	assert.Equal(t, "newuser@test.com", response["email"])
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "duplicate@test.com"
	password := "SecurePass123!"

	// Register first user
	ts.RegisterTestUser(t, email, password)

	// Try to register with same email
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/register", reqBody, nil)

	AssertErrorResponse(t, w, http.StatusConflict, "already exists")
}

func TestRegister_InvalidEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	reqBody := map[string]string{
		"email":    "invalid-email",
		"password": "SecurePass123!",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/register", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_WeakPassword(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	reqBody := map[string]string{
		"email":    "test@test.com",
		"password": "123",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/register", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "login@test.com"
	password := "SecurePass123!"

	// Register user first
	ts.RegisterTestUser(t, email, password)

	// Login
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/login", reqBody, nil)

	require.Equal(t, http.StatusOK, w.Code, "Failed to login: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
	assert.NotEmpty(t, response["expires_in"])
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "login_invalid@test.com"
	password := "SecurePass123!"

	// Register user
	ts.RegisterTestUser(t, email, password)

	// Try to login with wrong password
	reqBody := map[string]string{
		"email":    email,
		"password": "WrongPassword123!",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/login", reqBody, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "invalid")
}

func TestLogin_NonExistentUser(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	reqBody := map[string]string{
		"email":    "nonexistent@test.com",
		"password": "SecurePass123!",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/login", reqBody, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "invalid")
}

func TestRefresh_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and get tokens
	email := "refresh@test.com"
	password := "SecurePass123!"
	registerResp := ts.RegisterTestUser(t, email, password)

	refreshToken, ok := registerResp["refresh_token"].(string)
	require.True(t, ok, "Refresh token not found")

	// Wait a moment to ensure token timestamps differ
	time.Sleep(100 * time.Millisecond)

	// Refresh tokens
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/refresh", reqBody, nil)

	require.Equal(t, http.StatusOK, w.Code, "Failed to refresh: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	newAccessToken := response["access_token"].(string)
	newRefreshToken := response["refresh_token"].(string)

	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, refreshToken, newRefreshToken, "Refresh token should rotate")
}

func TestRefresh_InvalidToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	reqBody := map[string]string{
		"refresh_token": "invalid_token",
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/refresh", reqBody, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "invalid")
}

func TestRefresh_ExpiredToken(t *testing.T) {
	// This test would require mocking time or setting very short TTL
	// Skipped for now as it requires special test configuration
	t.Skip("Requires time mocking or very short TTL configuration")
}

func TestLogout_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login
	email := "logout@test.com"
	password := "SecurePass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Logout
	w := ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/auth/logout", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to logout: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "successfully")

	// Verify token is invalidated by trying to access protected endpoint
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_Unauthenticated(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/logout", nil, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "missing or invalid")
}

func TestLogoutAll_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register user
	email := "logoutall@test.com"
	password := "SecurePass123!"
	ts.RegisterTestUser(t, email, password)

	// Login multiple times to create multiple sessions
	token1 := ts.LoginTestUser(t, email, password)
	token2 := ts.LoginTestUser(t, email, password)
	token3 := ts.LoginTestUser(t, email, password)

	// All tokens should work
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token1)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token2)
	assert.Equal(t, http.StatusOK, w.Code)

	// Logout all sessions using token3
	w = ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/auth/logout-all", nil, token3)
	require.Equal(t, http.StatusOK, w.Code, "Failed to logout all: %s", w.Body.String())

	// All tokens should now be invalid
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token1)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token2)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token3)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_CompleteFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "complete@test.com"
	password := "SecurePass123!"

	// 1. Register
	registerResp := ts.RegisterTestUser(t, email, password)
	accessToken1 := registerResp["access_token"].(string)
	refreshToken1 := registerResp["refresh_token"].(string)

	// 2. Access protected resource with initial token
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken1)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Refresh tokens
	reqBody := map[string]string{
		"refresh_token": refreshToken1,
	}
	w = ts.MakeRequest(http.MethodPost, "/api/v1/auth/refresh", reqBody, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var refreshResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &refreshResp)
	require.NoError(t, err)

	accessToken2 := refreshResp["access_token"].(string)

	// 4. Access protected resource with new token
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken2)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Logout
	w = ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/auth/logout", nil, accessToken2)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Verify token is invalid after logout
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken2)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 7. Login again
	newToken := ts.LoginTestUser(t, email, password)

	// 8. Should be able to access resources again
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, newToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MultipleDevices(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "multidevice@test.com"
	password := "SecurePass123!"

	// Register once
	ts.RegisterTestUser(t, email, password)

	// Login from multiple "devices" (create multiple sessions)
	device1Token := ts.LoginTestUser(t, email, password)
	device2Token := ts.LoginTestUser(t, email, password)
	device3Token := ts.LoginTestUser(t, email, password)

	// All devices should have access
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device1Token)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device2Token)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device3Token)
	assert.Equal(t, http.StatusOK, w.Code)

	// Logout from device2
	w = ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/auth/logout", nil, device2Token)
	assert.Equal(t, http.StatusOK, w.Code)

	// Device1 and Device3 should still work
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device1Token)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device3Token)
	assert.Equal(t, http.StatusOK, w.Code)

	// Device2 should be logged out
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, device2Token)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_TokenReuse(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "tokenreuse@test.com"
	password := "SecurePass123!"

	// Register and get tokens
	registerResp := ts.RegisterTestUser(t, email, password)
	refreshToken := registerResp["refresh_token"].(string)

	// Use refresh token once
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/refresh", reqBody, nil)
	require.Equal(t, http.StatusOK, w.Code)

	// Try to reuse the same refresh token (should fail due to rotation)
	w = ts.MakeRequest(http.MethodPost, "/api/v1/auth/refresh", reqBody, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Refresh token should not be reusable after rotation")
}

func TestAuth_ConcurrentSessions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	email := "concurrent@test.com"
	password := "SecurePass123!"

	// Register user
	ts.RegisterTestUser(t, email, password)

	// Create 5 concurrent sessions
	tokens := make([]string, 5)
	for i := 0; i < 5; i++ {
		tokens[i] = ts.LoginTestUser(t, email, password)
	}

	// All sessions should be valid
	for i, token := range tokens {
		w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token)
		assert.Equal(t, http.StatusOK, w.Code, "Session %d should be valid", i+1)
	}

	// Logout all
	w := ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/auth/logout-all", nil, tokens[0])
	require.Equal(t, http.StatusOK, w.Code)

	// All sessions should be invalid
	for i, token := range tokens {
		w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, token)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Session %d should be invalid after logout-all", i+1)
	}
}
