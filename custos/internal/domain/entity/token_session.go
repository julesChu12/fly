package entity

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token for JWT rotation
type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	TokenHash string    `json:"-" gorm:"size:64;not null;index"` // SHA-256 hash
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID uint, token string, expiresAt time.Time) *RefreshToken {
	hash := sha256.Sum256([]byte(token))
	return &RefreshToken{
		UserID:    userID,
		TokenHash: base64.RawURLEncoding.EncodeToString(hash[:]),
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// MarkAsUsed marks the refresh token as used
func (rt *RefreshToken) MarkAsUsed() {
	rt.IsUsed = true
}

// Session represents a user session
type Session struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           uint      `json:"user_id" gorm:"not null;index"`
	SessionID        string    `json:"session_id" gorm:"size:36;not null;uniqueIndex"` // UUID
	RefreshTokenID   *uint     `json:"refresh_token_id,omitempty" gorm:"index"`
	DeviceID         string    `json:"device_id,omitempty" gorm:"size:128"`
	UserAgent        string    `json:"user_agent,omitempty" gorm:"size:500"`
	IP               string    `json:"ip,omitempty" gorm:"size:45"` // IPv4/IPv6
	Revoked          bool      `json:"revoked" gorm:"default:false"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	LastSeenAt       time.Time `json:"last_seen_at" gorm:"autoCreateTime"`

	// Relations
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	RefreshToken *RefreshToken `json:"refresh_token,omitempty" gorm:"foreignKey:RefreshTokenID"`
}

func (Session) TableName() string {
	return "sessions"
}

// NewSession creates a new session
func NewSession(userID uint, userAgent, ip string) *Session {
	sessionID := uuid.New().String()
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		UserAgent:  userAgent,
		IP:         ip,
		LastSeenAt: time.Now(),
	}
}

// Revoke revokes the session
func (s *Session) Revoke() {
	s.Revoked = true
}

// UpdateLastSeen updates the last seen timestamp
func (s *Session) UpdateLastSeen() {
	s.LastSeenAt = time.Now()
}

// IsValid checks if the session is valid (not revoked)
func (s *Session) IsValid() bool {
	return !s.Revoked
}

// JWKKey represents a JWK key for token signing/verification
type JWKKey struct {
	Kid       string     `json:"kid" gorm:"primaryKey;size:64"`
	Alg       string     `json:"alg" gorm:"size:16;not null"`
	PublicJWK string     `json:"public_jwk" gorm:"type:json;not null"`
	Active    bool       `json:"active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	RotatedAt *time.Time `json:"rotated_at,omitempty"`
	RetiredAt *time.Time `json:"retired_at,omitempty"`
}

func (JWKKey) TableName() string {
	return "jwk_keys"
}

// NewJWKKey creates a new JWK key
func NewJWKKey(kid, alg, publicJWK string) *JWKKey {
	return &JWKKey{
		Kid:       kid,
		Alg:       alg,
		PublicJWK: publicJWK,
		Active:    true,
	}
}

// Rotate marks the key as rotated
func (k *JWKKey) Rotate() {
	now := time.Now()
	k.RotatedAt = &now
	k.Active = false
}

// Retire marks the key as retired
func (k *JWKKey) Retire() {
	now := time.Now()
	k.RetiredAt = &now
	k.Active = false
}