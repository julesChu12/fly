package types

import "time"

type CreateCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Tags  string `json:"tags"`
}

type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Tags  string `json:"tags"`
}

type CustomerResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Contacts  []ContactResponse `json:"contacts,omitempty"`
}

type CreateContactRequest struct {
	CustomerID uint   `json:"customer_id" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Value      string `json:"value" binding:"required"`
	IsPrimary  bool   `json:"is_primary"`
}

type UpdateContactRequest struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	IsPrimary bool   `json:"is_primary"`
}

type ContactResponse struct {
	ID         uint      `json:"id"`
	CustomerID uint      `json:"customer_id"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type ListResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}