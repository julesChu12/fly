package entity

import "time"

// Session represents a persisted login session associated with a refresh token.
type Session struct {
	ID                    string     `json:"id" gorm:"primaryKey;size:72"`
	UserID                uint       `json:"user_id" gorm:"index;not null"`
	RefreshTokenHash      string     `json:"-" gorm:"size:128;not null"`
	RefreshTokenExpiresAt time.Time  `json:"refresh_token_expires_at" gorm:"not null"`
	IPAddress             string     `json:"ip_address" gorm:"size:45"`
	UserAgent             string     `json:"user_agent" gorm:"size:255"`
	LastUsedAt            time.Time  `json:"last_used_at" gorm:"not null"`
	RevokedAt             *time.Time `json:"revoked_at"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// IsActive reports whether the session is currently valid (not revoked and not expired).
func (s *Session) IsActive(now time.Time) bool {
	if s.RevokedAt != nil && !s.RevokedAt.IsZero() {
		return false
	}
	return now.Before(s.RefreshTokenExpiresAt)
}

// Revoke marks the session as revoked at the provided time.
func (s *Session) Revoke(at time.Time) {
	s.RevokedAt = &at
}

// Touch updates LastUsedAt to the supplied time.
func (s *Session) Touch(at time.Time) {
	s.LastUsedAt = at
}
