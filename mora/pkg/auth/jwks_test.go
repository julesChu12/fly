package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWKSValidator(t *testing.T) {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey
	keyID := "test-key-1"

	// Create mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := createMockJWKS(publicKey, keyID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	tests := []struct {
		name      string
		setupFunc func() (string, error)
		wantErr   bool
		errType   error
	}{
		{
			name: "valid token with JWKS",
			setupFunc: func() (string, error) {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				token.Header["kid"] = keyID
				return token.SignedString(privateKey)
			},
			wantErr: false,
		},
		{
			name: "token without kid header",
			setupFunc: func() (string, error) {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				// Don't set kid header
				return token.SignedString(privateKey)
			},
			wantErr: true,
		},
		{
			name: "token with unknown kid",
			setupFunc: func() (string, error) {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				token.Header["kid"] = "unknown-key"
				return token.SignedString(privateKey)
			},
			wantErr: true,
		},
		{
			name: "expired token",
			setupFunc: func() (string, error) {
				claims := NewClaims("user-123", "testuser", -1*time.Hour) // Expired
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				token.Header["kid"] = keyID
				return token.SignedString(privateKey)
			},
			wantErr: true,
			errType: ErrExpiredToken,
		},
	}

	validator := NewJWKSValidator(jwksServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Failed to create test token: %v", err)
			}

			claims, err := validator.ValidateTokenWithJWKS(tokenString)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if claims == nil {
				t.Error("Expected claims, got nil")
				return
			}

			if claims.UserID != "user-123" {
				t.Errorf("Expected UserID 'user-123', got %s", claims.UserID)
			}

			if claims.Username != "testuser" {
				t.Errorf("Expected Username 'testuser', got %s", claims.Username)
			}
		})
	}
}

func TestValidateTokenWithPublicKey(t *testing.T) {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Convert public key to PEM format
	publicKeyPEM, err := publicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to convert public key to PEM: %v", err)
	}

	tests := []struct {
		name         string
		tokenSetup   func() string
		publicKeyPEM string
		wantErr      bool
		errType      error
	}{
		{
			name: "valid token with public key",
			tokenSetup: func() string {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			publicKeyPEM: publicKeyPEM,
			wantErr:      false,
		},
		{
			name: "invalid public key",
			tokenSetup: func() string {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			publicKeyPEM: "invalid-pem",
			wantErr:      true,
		},
		{
			name: "empty token",
			tokenSetup: func() string {
				return ""
			},
			publicKeyPEM: publicKeyPEM,
			wantErr:      true,
			errType:      ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.tokenSetup()
			claims, err := ValidateTokenWithPublicKey(tokenString, tt.publicKeyPEM)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if claims == nil {
				t.Error("Expected claims, got nil")
				return
			}

			if claims.UserID != "user-123" {
				t.Errorf("Expected UserID 'user-123', got %s", claims.UserID)
			}
		})
	}
}

func TestGenerateTokenWithPrivateKey(t *testing.T) {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Convert private key to PEM format
	privateKeyPEM, err := privateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("Failed to convert private key to PEM: %v", err)
	}

	tests := []struct {
		name          string
		userID        string
		username      string
		privateKeyPEM string
		ttl           time.Duration
		wantErr       bool
	}{
		{
			name:          "valid token generation",
			userID:        "user-123",
			username:      "testuser",
			privateKeyPEM: privateKeyPEM,
			ttl:           10 * time.Minute,
			wantErr:       false,
		},
		{
			name:          "invalid private key",
			userID:        "user-123",
			username:      "testuser",
			privateKeyPEM: "invalid-pem",
			ttl:           10 * time.Minute,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GenerateTokenWithPrivateKey(tt.userID, tt.username, tt.privateKeyPEM, tt.ttl)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tokenString == "" {
				t.Error("Expected token string, got empty")
				return
			}

			// Validate the generated token
			publicKeyPEM, err := publicKeyToPEM(&privateKey.PublicKey)
			if err != nil {
				t.Fatalf("Failed to convert public key to PEM: %v", err)
			}

			claims, err := ValidateTokenWithPublicKey(tokenString, publicKeyPEM)
			if err != nil {
				t.Errorf("Failed to validate generated token: %v", err)
				return
			}

			if claims.UserID != tt.userID {
				t.Errorf("Expected UserID %s, got %s", tt.userID, claims.UserID)
			}

			if claims.Username != tt.username {
				t.Errorf("Expected Username %s, got %s", tt.username, claims.Username)
			}
		})
	}
}

func TestGetKeyIDFromToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	tests := []struct {
		name        string
		tokenSetup  func() string
		expectedKID string
		wantErr     bool
	}{
		{
			name: "token with kid header",
			tokenSetup: func() string {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				token.Header["kid"] = "test-key-1"
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			expectedKID: "test-key-1",
			wantErr:     false,
		},
		{
			name: "token without kid header",
			tokenSetup: func() string {
				claims := NewClaims("user-123", "testuser", 10*time.Minute)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "malformed token",
			tokenSetup: func() string {
				return "invalid.token"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.tokenSetup()
			kid, err := GetKeyIDFromToken(tokenString)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if kid != tt.expectedKID {
				t.Errorf("Expected KID %s, got %s", tt.expectedKID, kid)
			}
		})
	}
}

// Helper functions for testing

func createMockJWKS(publicKey *rsa.PublicKey, keyID string) *JWKS {
	return &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: keyID,
				Use: "sig",
				Alg: "RS256",
				N:   encodeBase64URL(publicKey.N.Bytes()),
				E:   encodeBase64URL(big.NewInt(int64(publicKey.E)).Bytes()),
			},
		},
	}
}

func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func privateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

func publicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}
