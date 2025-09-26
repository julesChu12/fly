package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateToken(t *testing.T) {
	svc := NewTokenService("secret", time.Minute, time.Hour)

	pair, err := svc.GenerateAccessToken("session-1", 42, "alice", "admin")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.Equal(t, "Bearer", pair.TokenType)
	require.Equal(t, "session-1", pair.SessionID)

	claims, err := svc.ValidateToken(pair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, uint(42), claims.UserID)
	require.Equal(t, "alice", claims.Username)
	require.Equal(t, "admin", string(claims.Role))
}

func TestValidateTokenExpiry(t *testing.T) {
	svc := NewTokenService("secret", time.Millisecond, time.Hour)
	pair, err := svc.GenerateAccessToken("session-2", 1, "bob", "user")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	_, err = svc.ValidateToken(pair.AccessToken)
	require.Error(t, err)
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := NewTokenService("secret", time.Minute, time.Minute)

	refresh, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, refresh.Token)
	require.True(t, refresh.ExpiresAt.After(time.Now()))
	require.Equal(t, int64(time.Minute.Seconds()), refresh.ExpiresIn)

	hash := svc.HashRefreshToken(refresh.Token)
	require.NotEmpty(t, hash)
}
