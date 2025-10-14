package entity

import (
	"time"
)

type Customer struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  uint      `json:"tenant_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Phone     string    `json:"phone" gorm:"type:varchar(20);index"`
	Email     string    `json:"email" gorm:"type:varchar(255);index"`
	Tags      string    `json:"tags" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Contacts []Contact `json:"contacts" gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
}

type Contact struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID   uint      `json:"tenant_id" gorm:"not null;index"`
	CustomerID uint      `json:"customer_id" gorm:"not null;index"`
	Type       string    `json:"type" gorm:"type:varchar(50);not null"`
	Value      string    `json:"value" gorm:"type:varchar(255);not null"`
	IsPrimary  bool      `json:"is_primary" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Customer Customer `json:"customer" gorm:"foreignKey:CustomerID"`
}

func (Customer) TableName() string {
	return "customers"
}

func (Contact) TableName() string {
	return "contacts"
}