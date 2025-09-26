package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/julesChu12/fly/custos/pkg/constants"
	"github.com/julesChu12/fly/custos/pkg/errors"
	"github.com/julesChu12/fly/custos/pkg/types"
)

type TokenService struct {
	secretKey  string
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type TokenClaims struct {
	UserID    uint           `json:"user_id"`
	Username  string         `json:"username"`
	Role      types.UserRole `json:"role"`
	SessionID string         `json:"session_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	SessionID        string `json:"session_id"`
}

type RefreshToken struct {
	Token     string
	ExpiresAt time.Time
	ExpiresIn int64
}

func NewTokenService(secretKey string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secretKey:  secretKey,
		issuer:     constants.JWTIssuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *TokenService) GenerateAccessToken(sessionID string, userID uint, username string, role types.UserRole) (*TokenPair, error) {
	now := time.Now()
	claims := &TokenClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &TokenPair{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.accessTTL.Seconds()),
		SessionID:   sessionID,
	}, nil
}

func (s *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		// Check if token is expired
		if err.Error() == "token is expired" {
			return nil, errors.NewTokenExpiredError()
		}
		return nil, errors.NewTokenInvalidError()
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.NewTokenInvalidError()
	}

	return claims, nil
}

// GenerateRefreshToken produces a cryptographically secure refresh token string and expiry metadata.
func (s *TokenService) GenerateRefreshToken() (*RefreshToken, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	expiresAt := time.Now().Add(s.refreshTTL)
	return &RefreshToken{
		Token:     token,
		ExpiresAt: expiresAt,
		ExpiresIn: int64(s.refreshTTL.Seconds()),
	}, nil
}

// HashRefreshToken hashes the refresh token for safe persistence.
func (s *TokenService) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// GenerateSessionID creates a unique identifier for session records.
func (s *TokenService) GenerateSessionID() string {
	return uuid.NewString()
}

// RefreshTTL returns the configured refresh token duration.
func (s *TokenService) RefreshTTL() time.Duration {
	return s.refreshTTL
}
