package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		secret   string
		ttl      time.Duration
		wantErr  bool
	}{
		{
			name:     "valid token generation",
			userID:   "user123",
			username: "testuser",
			secret:   "test-secret",
			ttl:      time.Hour,
			wantErr:  false,
		},
		{
			name:     "empty secret",
			userID:   "user123",
			username: "testuser",
			secret:   "",
			ttl:      time.Hour,
			wantErr:  false, // HMAC can work with empty secret
		},
		{
			name:     "zero ttl",
			userID:   "user123",
			username: "testuser",
			secret:   "test-secret",
			ttl:      0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username, tt.secret, tt.ttl)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	userID := "user123"
	username := "testuser"
	ttl := time.Hour

	// Generate a valid token for testing
	validToken, err := GenerateToken(userID, username, secret, ttl)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Generate an expired token
	expiredToken, err := GenerateToken(userID, username, secret, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired test token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   error
		wantClaims bool
	}{
		{
			name:       "valid token",
			token:      validToken,
			secret:     secret,
			wantErr:    nil,
			wantClaims: true,
		},
		{
			name:       "empty token",
			token:      "",
			secret:     secret,
			wantErr:    ErrInvalidToken,
			wantClaims: false,
		},
		{
			name:       "invalid secret",
			token:      validToken,
			secret:     "wrong-secret",
			wantErr:    ErrInvalidToken,
			wantClaims: false,
		},
		{
			name:       "expired token",
			token:      expiredToken,
			secret:     secret,
			wantErr:    ErrExpiredToken,
			wantClaims: false,
		},
		{
			name:       "malformed token",
			token:      "invalid.token.format",
			secret:     secret,
			wantErr:    ErrMalformedToken,
			wantClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				if claims != nil {
					t.Error("ValidateToken() should return nil claims on error")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
				return
			}

			if tt.wantClaims {
				if claims == nil {
					t.Error("ValidateToken() returned nil claims")
					return
				}

				if claims.UserID != userID {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
				}

				if claims.Username != username {
					t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, username)
				}
			}
		})
	}
}

func TestClaimsIsExpired(t *testing.T) {
	tests := []struct {
		name       string
		claims     *Claims
		wantExpired bool
	}{
		{
			name: "not expired",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			},
			wantExpired: false,
		},
		{
			name: "expired",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				},
			},
			wantExpired: true,
		},
		{
			name: "no expiration",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{},
			},
			wantExpired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.claims.IsExpired(); got != tt.wantExpired {
				t.Errorf("Claims.IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

func TestNewClaims(t *testing.T) {
	userID := "user123"
	username := "testuser"
	ttl := time.Hour

	claims := NewClaims(userID, username, ttl)

	if claims.UserID != userID {
		t.Errorf("NewClaims() UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("NewClaims() Username = %v, want %v", claims.Username, username)
	}

	if claims.Subject != userID {
		t.Errorf("NewClaims() Subject = %v, want %v", claims.Subject, userID)
	}

	if claims.ExpiresAt == nil {
		t.Error("NewClaims() ExpiresAt should not be nil")
	}

	if claims.IssuedAt == nil {
		t.Error("NewClaims() IssuedAt should not be nil")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	secret := "test-secret-for-roundtrip"
	userID := "user456"
	username := "roundtripuser"
	ttl := 30 * time.Minute

	// Generate token
	token, err := GenerateToken(userID, username, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	// Validate token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}

	// Verify claims
	if claims.UserID != userID {
		t.Errorf("Round trip UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("Round trip Username = %v, want %v", claims.Username, username)
	}

	if claims.IsExpired() {
		t.Error("Token should not be expired immediately after generation")
	}
}