package entity

import (
	"time"
)

// Tenant represents a tenant/organization in the multi-tenant system
type Tenant struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;size:50;not null"` // URL-friendly identifier
	Domain      string    `json:"domain" gorm:"uniqueIndex;size:100"`       // Custom domain (optional)
	Status      string    `json:"status" gorm:"size:20;not null;default:'active'"`
	Plan        string    `json:"plan" gorm:"size:20;not null;default:'basic'"` // basic, premium, enterprise
	Settings    string    `json:"settings" gorm:"type:json"`                     // JSON settings
	MaxUsers    int       `json:"max_users" gorm:"default:10"`                   // User limit
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Users       []User     `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	// Note: Customer and Service entities will be defined in future business modules
}

// TenantSettings represents tenant configuration
type TenantSettings struct {
	BusinessType    string            `json:"business_type"`    // hair_salon, car_wash, restaurant, etc.
	Timezone        string            `json:"timezone"`
	Currency        string            `json:"currency"`
	Language        string            `json:"language"`
	BusinessHours   map[string]string `json:"business_hours"`   // day -> hours
	ContactInfo     ContactInfo       `json:"contact_info"`
	BrandingColors  BrandingColors    `json:"branding_colors"`
	Features        []string          `json:"features"`         // enabled features
}

type ContactInfo struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Website  string `json:"website"`
}

type BrandingColors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Accent    string `json:"accent"`
}

// NewTenant creates a new tenant
func NewTenant(name, slug string) *Tenant {
	return &Tenant{
		Name:      name,
		Slug:      slug,
		Status:    "active",
		Plan:      "basic",
		MaxUsers:  10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsActive checks if tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == "active"
}

// CanAddUser checks if tenant can add more users
func (t *Tenant) CanAddUser(currentUserCount int) bool {
	return currentUserCount < t.MaxUsers
}

// GetDomain returns the tenant's domain (custom or default)
func (t *Tenant) GetDomain() string {
	if t.Domain != "" {
		return t.Domain
	}
	return t.Slug + ".fly.local" // default subdomain
}