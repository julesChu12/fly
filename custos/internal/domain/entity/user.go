package entity

import (
	"time"

	"github.com/julesChu12/custos/pkg/types"
)

type User struct {
	ID        uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string           `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email     string           `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password  string           `json:"-" gorm:"size:255;not null"`
	Nickname  string           `json:"nickname" gorm:"size:100"`
	Avatar    string           `json:"avatar" gorm:"size:255"`
	Status    types.UserStatus `json:"status" gorm:"size:20;not null;default:'active'"`
	Role      types.UserRole   `json:"role" gorm:"size:20;not null;default:'user'"`
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

func NewUser(username, email, hashedPassword string) *User {
	return &User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Status:   types.UserStatusActive,
		Role:     types.UserRoleUser,
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

func (u *User) SetOAuthProvider(provider, providerID string) {
	// TODO: Add OAuth provider fields to User entity
	// For now, this is a placeholder
}
