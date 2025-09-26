package entity

import (
	"time"
)

// UserProfile represents extended user profile information
type UserProfile struct {
	UserID    uint       `json:"user_id" gorm:"primaryKey"`
	Nickname  string     `json:"nickname" gorm:"size:64"`
	Avatar    string     `json:"avatar" gorm:"size:255"`
	Gender    string     `json:"gender" gorm:"type:enum('male','female','other');default:'other'"`
	Birthday  *time.Time `json:"birthday,omitempty" gorm:"type:date"`
	Extra     string     `json:"extra,omitempty" gorm:"type:json"` // JSON for additional fields
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// NewUserProfile creates a new user profile
func NewUserProfile(userID uint) *UserProfile {
	return &UserProfile{
		UserID: userID,
		Gender: "other",
	}
}

// UpdateProfile updates profile information
func (up *UserProfile) UpdateProfile(nickname, avatar string, birthday *time.Time) {
	if nickname != "" {
		up.Nickname = nickname
	}
	if avatar != "" {
		up.Avatar = avatar
	}
	if birthday != nil {
		up.Birthday = birthday
	}
}