package entity

import (
	"time"
)

// UserOAuth represents OAuth binding between user and external provider
type UserOAuth struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	Provider     string    `json:"provider" gorm:"size:64;not null"` // google/github/wechat
	ProviderUID  string    `json:"provider_uid" gorm:"size:128;not null"`
	AccessToken  string    `json:"-" gorm:"size:255"`
	RefreshToken string    `json:"-" gorm:"size:255"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserOAuth) TableName() string {
	return "user_oauth"
}

// NewUserOAuth creates a new OAuth binding
func NewUserOAuth(userID uint, provider, providerUID string) *UserOAuth {
	return &UserOAuth{
		UserID:      userID,
		Provider:    provider,
		ProviderUID: providerUID,
	}
}

// IsExpired checks if the OAuth token is expired
func (uo *UserOAuth) IsExpired() bool {
	if uo.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*uo.ExpiresAt)
}

// UpdateTokens updates access and refresh tokens
func (uo *UserOAuth) UpdateTokens(accessToken, refreshToken string, expiresAt *time.Time) {
	uo.AccessToken = accessToken
	uo.RefreshToken = refreshToken
	uo.ExpiresAt = expiresAt
}