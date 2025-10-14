package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfile_AutoCreate(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_autocreate@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Get profile (should auto-create)
	w := ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to get profile: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify default profile was created
	assert.NotNil(t, response["user_id"])
	assert.Equal(t, "", response["nickname"])
	assert.Equal(t, "", response["avatar"])
	assert.Equal(t, "", response["gender"])
}

func TestGetProfile_Unauthenticated(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to get profile without authentication
	w := ts.MakeRequest(http.MethodGet, "/api/v1/profile", nil, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "missing or invalid")
}

func TestUpdateProfile_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_update@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Update profile
	updateReq := map[string]interface{}{
		"nickname": "TestUser",
		"avatar":   "https://example.com/avatar.jpg",
		"gender":   "male",
		"birthday": "1990-01-15",
		"extra":    "{\"hobby\": \"coding\"}",
	}

	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", updateReq, accessToken)

	require.Equal(t, http.StatusOK, w.Code, "Failed to update profile: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "profile updated successfully", response["message"])

	// Verify the update by getting the profile
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "TestUser", response["nickname"])
	assert.Equal(t, "https://example.com/avatar.jpg", response["avatar"])
	assert.Equal(t, "male", response["gender"])
	assert.Equal(t, "1990-01-15", response["birthday"])
	assert.Equal(t, "{\"hobby\": \"coding\"}", response["extra"])
}

func TestUpdateProfile_PartialUpdate(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_partial@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// First update with full data
	fullUpdate := map[string]interface{}{
		"nickname": "OriginalName",
		"avatar":   "https://example.com/original.jpg",
		"gender":   "female",
		"birthday": "1995-05-20",
	}

	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", fullUpdate, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Partial update (only nickname)
	partialUpdate := map[string]interface{}{
		"nickname": "UpdatedName",
	}

	w = ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", partialUpdate, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Verify partial update kept other fields
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "UpdatedName", response["nickname"])
	assert.Equal(t, "https://example.com/original.jpg", response["avatar"]) // Should remain unchanged
	assert.Equal(t, "female", response["gender"])                           // Should remain unchanged
	assert.Equal(t, "1995-05-20", response["birthday"])                     // Should remain unchanged
}

func TestUpdateProfile_InvalidBirthdayFormat(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_invalid_birthday@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Try to update with invalid birthday format
	updateReq := map[string]interface{}{
		"birthday": "01/15/1990", // Invalid format, should be YYYY-MM-DD
	}

	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", updateReq, accessToken)

	AssertErrorResponse(t, w, http.StatusBadRequest, "invalid birthday format")
}

func TestUpdateProfile_InvalidGender(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_invalid_gender@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Try to update with invalid gender
	updateReq := map[string]interface{}{
		"gender": "invalid_gender",
	}

	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", updateReq, accessToken)

	// Should fail validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfile_MaxLengthValidation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_maxlength@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// Try to update with nickname exceeding max length (64 chars)
	longNickname := string(make([]byte, 65))
	for i := range longNickname {
		longNickname = longNickname[:i] + "a" + longNickname[i+1:]
	}

	updateReq := map[string]interface{}{
		"nickname": longNickname,
	}

	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", updateReq, accessToken)

	// Should fail validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfile_Unauthenticated(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to update profile without authentication
	updateReq := map[string]interface{}{
		"nickname": "TestUser",
	}

	w := ts.MakeRequest(http.MethodPut, "/api/v1/profile", updateReq, nil)

	AssertErrorResponse(t, w, http.StatusUnauthorized, "missing or invalid")
}

func TestProfile_MultipleUsersIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create two users
	user1Token := ts.RegisterAndLoginTestUser(t, "user1@test.com", "Pass123!")
	user2Token := ts.RegisterAndLoginTestUser(t, "user2@test.com", "Pass123!")

	// Update user1 profile
	user1Update := map[string]interface{}{
		"nickname": "User1",
		"gender":   "male",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", user1Update, user1Token)
	require.Equal(t, http.StatusOK, w.Code)

	// Update user2 profile
	user2Update := map[string]interface{}{
		"nickname": "User2",
		"gender":   "female",
	}
	w = ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", user2Update, user2Token)
	require.Equal(t, http.StatusOK, w.Code)

	// Verify user1 profile
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, user1Token)
	require.Equal(t, http.StatusOK, w.Code)

	var user1Profile map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user1Profile)
	require.NoError(t, err)
	assert.Equal(t, "User1", user1Profile["nickname"])
	assert.Equal(t, "male", user1Profile["gender"])

	// Verify user2 profile
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, user2Token)
	require.Equal(t, http.StatusOK, w.Code)

	var user2Profile map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &user2Profile)
	require.NoError(t, err)
	assert.Equal(t, "User2", user2Profile["nickname"])
	assert.Equal(t, "female", user2Profile["gender"])

	// Ensure profiles are isolated
	assert.NotEqual(t, user1Profile["user_id"], user2Profile["user_id"])
}

func TestProfile_EmptyFieldsUpdate(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Register and login a user
	email := "profile_empty@test.com"
	password := "TestPass123!"
	accessToken := ts.RegisterAndLoginTestUser(t, email, password)

	// First set some data
	initialUpdate := map[string]interface{}{
		"nickname": "InitialName",
		"avatar":   "https://example.com/avatar.jpg",
		"gender":   "male",
	}
	w := ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", initialUpdate, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Try to update with empty values (should be ignored due to omitempty)
	emptyUpdate := map[string]interface{}{}
	w = ts.MakeAuthenticatedRequest(http.MethodPut, "/api/v1/profile", emptyUpdate, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Verify original data is preserved
	w = ts.MakeAuthenticatedRequest(http.MethodGet, "/api/v1/profile", nil, accessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "InitialName", response["nickname"])
	assert.Equal(t, "https://example.com/avatar.jpg", response["avatar"])
	assert.Equal(t, "male", response["gender"])
}
