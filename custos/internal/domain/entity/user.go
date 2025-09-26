package entity

import (
	"time"

	"github.com/julesChu12/fly/custos/pkg/types"
)

type User struct {
	ID                  uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Username            string           `json:"username" gorm:"uniqueIndex;size:50"`
	Email               string           `json:"email" gorm:"uniqueIndex;size:100"`
	Password            string           `json:"-" gorm:"size:255"`
	Nickname            string           `json:"nickname" gorm:"size:100"`
	Avatar              string           `json:"avatar" gorm:"size:255"`
	Status              types.UserStatus `json:"status" gorm:"size:20;not null;default:'active'"`
	Role                types.UserRole   `json:"role" gorm:"size:20;not null;default:'user'"`
	UserType            types.UserType   `json:"user_type" gorm:"size:20;default:'customer'"`
	TenantID            *uint            `json:"tenant_id,omitempty" gorm:"index"`
	TokenVersion        int              `json:"token_version" gorm:"default:0;index"`
	MergedIntoUserID    *uint            `json:"merged_into_user_id,omitempty"`
	LastLoginAt         *time.Time       `json:"last_login_at,omitempty"`
	CreatedAt           time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time        `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	OAuthBindings       []UserOAuth      `json:"oauth_bindings,omitempty" gorm:"foreignKey:UserID"`
	Profile             *UserProfile     `json:"profile,omitempty" gorm:"foreignKey:UserID"`
	Sessions            []Session        `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	RefreshTokens       []RefreshToken   `json:"refresh_tokens,omitempty" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

func NewUser(username, email, hashedPassword string) *User {
	return &User{
		Username:     username,
		Email:        email,
		Password:     hashedPassword,
		Status:       types.UserStatusActive,
		Role:         types.UserRoleUser,
		UserType:     types.UserTypeCustomer,
		TokenVersion: 0,
	}
}

func (u *User) IsActive() bool {
	return u.Status == types.UserStatusActive
}

func (u *User) IsAdmin() bool {
	return u.Role == types.UserRoleAdmin
}

func (u *User) Activate() {
	u.Status = types.UserStatusActive
}

func (u *User) Deactivate() {
	u.Status = types.UserStatusInactive
}

func (u *User) SetLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

func (u *User) IncrementTokenVersion() {
	u.TokenVersion++
}

func (u *User) IsTokenVersionValid(version int) bool {
	return u.TokenVersion == version
}

func (u *User) MergeInto(targetUserID uint) {
	u.Status = types.UserStatusMerged
	u.MergedIntoUserID = &targetUserID
}

// SetOAuthProvider sets OAuth provider information for the user
func (u *User) SetOAuthProvider(provider, providerID string) {
	// This method can be used to update or create OAuth bindings
	// Implementation depends on how OAuth bindings are managed
	// For now, this is a placeholder that can be expanded later
}
