package entity

import (
	"encoding/json"
	"time"
)

// UserProfile represents extended user profile information
type UserProfile struct {
	UserID    uint       `json:"user_id" gorm:"primaryKey"`
	Nickname  string     `json:"nickname" gorm:"size:64"`
	Avatar    string     `json:"avatar" gorm:"size:255"`
	Gender    string     `json:"gender" gorm:"type:enum('male','female','other');default:'other'"`
	Birthday  *time.Time `json:"birthday,omitempty" gorm:"type:date"`
	Extra     *string    `json:"extra,omitempty" gorm:"type:json"` // JSON for additional fields (pointer to support NULL)
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// NewUserProfile creates a new user profile with default values
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

// SetGender sets the gender with validation
func (up *UserProfile) SetGender(gender string) error {
	validGenders := map[string]bool{
		"male":   true,
		"female": true,
		"other":  true,
	}

	if !validGenders[gender] {
		return ErrInvalidGender
	}

	up.Gender = gender
	return nil
}

// SetExtra sets extra data as JSON
func (up *UserProfile) SetExtra(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	jsonStr := string(jsonData)
	up.Extra = &jsonStr
	return nil
}

// GetExtra retrieves extra data from JSON
func (up *UserProfile) GetExtra(v interface{}) error {
	if up.Extra == nil || *up.Extra == "" {
		return nil
	}
	return json.Unmarshal([]byte(*up.Extra), v)
}

// IsComplete checks if the profile has all basic information filled
func (up *UserProfile) IsComplete() bool {
	return up.Nickname != "" && up.Avatar != ""
}

// Age calculates the age from birthday
func (up *UserProfile) Age() int {
	if up.Birthday == nil {
		return 0
	}

	now := time.Now()
	years := now.Year() - up.Birthday.Year()

	// Adjust if birthday hasn't occurred this year
	if now.Month() < up.Birthday.Month() ||
		(now.Month() == up.Birthday.Month() && now.Day() < up.Birthday.Day()) {
		years--
	}

	return years
}
