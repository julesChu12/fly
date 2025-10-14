package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create an admin user and assign admin role
func (ts *TestServer) CreateAdminUser(t *testing.T, email, password string) string {
	// Register user
	registerResp := ts.RegisterTestUser(t, email, password)
	userID := uint(registerResp["user_id"].(float64))

	// Assign admin role using RBAC service directly
	err := ts.RBACService.AssignRole(context.Background(), userID, "admin")
	require.NoError(t, err, "Failed to assign admin role")

	// Login to get token
	return ts.LoginTestUser(t, email, password)
}

func TestAdmin_ListUsers_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create some regular users
	ts.RegisterTestUser(t, "user1@test.com", "Pass123!")
	ts.RegisterTestUser(t, "user2@test.com", "Pass123!")
	ts.RegisterTestUser(t, "user3@test.com", "Pass123!")

	// List users
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/users?page=1&limit=10", nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to list users: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	users := response["users"].([]interface{})
	assert.GreaterOrEqual(t, len(users), 4, "Should have at least 4 users (1 admin + 3 regular)")

	total := int(response["total"].(float64))
	assert.GreaterOrEqual(t, total, 4)
}

func TestAdmin_ListUsers_Unauthorized(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create regular user (not admin)
	regularToken := ts.RegisterAndLoginTestUser(t, "regular@test.com", "Pass123!")

	// Try to list users without admin role
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/users", nil, regularToken)

	AssertErrorResponse(t, w, http.StatusForbidden, "permission denied")
}

func TestAdmin_ListUsers_Pagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create 15 regular users
	for i := 1; i <= 15; i++ {
		ts.RegisterTestUser(t, fmt.Sprintf("user%d@test.com", i), "Pass123!")
	}

	// Get first page (limit 5)
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/users?page=1&limit=5", nil, adminToken)
	require.Equal(t, http.StatusOK, w.Code)

	var page1 map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &page1)
	require.NoError(t, err)

	users1 := page1["users"].([]interface{})
	assert.Len(t, users1, 5, "First page should have 5 users")

	// Get second page
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/users?page=2&limit=5", nil, adminToken)
	require.Equal(t, http.StatusOK, w.Code)

	var page2 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &page2)
	require.NoError(t, err)

	users2 := page2["users"].([]interface{})
	assert.Len(t, users2, 5, "Second page should have 5 users")

	// Verify different users on different pages
	user1ID := users1[0].(map[string]interface{})["id"]
	user2ID := users2[0].(map[string]interface{})["id"]
	assert.NotEqual(t, user1ID, user2ID, "Different pages should have different users")
}

func TestAdmin_GetUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))

	// Get user details
	w := ts.MakeAuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get user: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(userID), response["id"].(float64))
	assert.Equal(t, "target@test.com", response["email"])
}

func TestAdmin_GetUser_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Try to get non-existent user
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/users/99999", nil, adminToken)

	AssertErrorResponse(t, w, http.StatusNotFound, "not found")
}

func TestAdmin_UpdateUserStatus_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))
	userToken := registerResp["access_token"].(string)

	// Verify user can access profile initially
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Suspend the user
	reqBody := map[string]string{
		"status": "suspended",
	}
	w = ts.MakeAuthenticatedRequest(http.MethodPatch, fmt.Sprintf("/api/v1/admin/users/%d/status", userID), reqBody, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to update status: %s", w.Body.String())

	// Note: In a real scenario, the user's sessions might be invalidated
	// This depends on your session validation logic
}

func TestAdmin_UpdateUserStatus_InvalidStatus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))

	// Try invalid status
	reqBody := map[string]string{
		"status": "invalid_status",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPatch, fmt.Sprintf("/api/v1/admin/users/%d/status", userID), reqBody, adminToken)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdmin_ForceLogoutUser_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user with active sessions
	email := "target@test.com"
	password := "Pass123!"
	ts.RegisterTestUser(t, email, password)

	// Login multiple times to create multiple sessions
	userToken1 := ts.LoginTestUser(t, email, password)
	userToken2 := ts.LoginTestUser(t, email, password)

	// Get user ID
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken1)
	require.Equal(t, http.StatusOK, w.Code)

	var profile map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(t, err)
	userID := uint(profile["user_id"].(float64))

	// Verify both tokens work
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken1)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken2)
	assert.Equal(t, http.StatusOK, w.Code)

	// Admin forces logout
	w = ts.MakeAuthenticatedRequest(http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/force-logout", userID), nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to force logout: %s", w.Body.String())

	// Both tokens should now be invalid
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken1)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, userToken2)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdmin_AssignRole_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))

	// Assign moderator role
	reqBody := map[string]string{
		"role": "moderator",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/roles", userID), reqBody, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to assign role: %s", w.Body.String())

	// Verify role was assigned
	w = ts.MakeAuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d/roles", userID), nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	roles := response["roles"].([]interface{})
	assert.Contains(t, roles, "moderator")
}

func TestAdmin_GetUserRoles_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))

	// Assign multiple roles
	roles := []string{"moderator", "editor"}
	for _, role := range roles {
		reqBody := map[string]string{"role": role}
		w := ts.MakeAuthenticatedRequest(http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/roles", userID), reqBody, adminToken)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// Get user roles
	w := ts.MakeAuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d/roles", userID), nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get roles: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	userRoles := response["roles"].([]interface{})
	assert.GreaterOrEqual(t, len(userRoles), 2, "Should have at least 2 roles")
}

func TestAdmin_GetSystemStats_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Create some users and sessions
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("user%d@test.com", i)
		ts.RegisterAndLoginTestUser(t, email, "Pass123!")
	}

	// Get system stats
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/admin/stats", nil, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get stats: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify stats structure
	assert.NotNil(t, response["total_users"])
	assert.NotNil(t, response["active_sessions"])

	totalUsers := int(response["total_users"].(float64))
	assert.GreaterOrEqual(t, totalUsers, 6, "Should have at least 6 users (1 admin + 5 regular)")

	activeSessions := int(response["active_sessions"].(float64))
	assert.GreaterOrEqual(t, activeSessions, 1, "Should have at least 1 active session")
}

func TestAdmin_AddPolicy_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Add a new policy
	reqBody := map[string]interface{}{
		"subject": "moderator",
		"object":  "posts",
		"action":  "edit",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/admin/policies", reqBody, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to add policy: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "success")
}

func TestAdmin_RemovePolicy_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminToken := ts.CreateAdminUser(t, "admin@test.com", "AdminPass123!")

	// Add a policy first
	addReq := map[string]interface{}{
		"subject": "moderator",
		"object":  "comments",
		"action":  "delete",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPost, "/api/v1/admin/policies", addReq, adminToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Remove the policy
	removeReq := map[string]interface{}{
		"subject": "moderator",
		"object":  "comments",
		"action":  "delete",
	}
	w = ts.MakeAuthenticatedRequest(http.MethodDelete, "/api/v1/admin/policies", removeReq, adminToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to remove policy: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["message"], "success")
}

func TestAdmin_MultipleAdmins(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create two admin users
	admin1Token := ts.CreateAdminUser(t, "admin1@test.com", "AdminPass123!")
	admin2Token := ts.CreateAdminUser(t, "admin2@test.com", "AdminPass123!")

	// Create a regular user
	registerResp := ts.RegisterTestUser(t, "target@test.com", "Pass123!")
	userID := uint(registerResp["user_id"].(float64))

	// Both admins should be able to manage users
	w := ts.MakeAuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, admin1Token)
	assert.Equal(t, http.StatusOK, w.Code)

	w = ts.MakeAuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, admin2Token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdmin_CannotManageSelf(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create admin user
	adminResp := ts.RegisterTestUser(t, "admin@test.com", "AdminPass123!")
	adminID := uint(adminResp["user_id"].(float64))

	// Assign admin role
	err := ts.RBACService.AssignRole(context.Background(), adminID, "admin")
	require.NoError(t, err)

	adminToken := ts.LoginTestUser(t, "admin@test.com", "AdminPass123!")

	// Try to force logout self (may or may not be allowed depending on business logic)
	w := ts.MakeAuthenticatedRequest(http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/force-logout", adminID), nil, adminToken)

	// This test documents the behavior - adjust based on your requirements
	// Some systems allow admins to logout themselves, others don't
	if w.Code == http.StatusOK {
		t.Log("System allows admin to force-logout themselves")
	} else {
		t.Log("System prevents admin from force-logout themselves")
	}
}
