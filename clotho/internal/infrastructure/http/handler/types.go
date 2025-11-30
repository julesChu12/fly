package handler

import "time"

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error      string    `json:"error" example:"invalid_request"`
	Message    string    `json:"message" example:"The request is invalid"`
	Timestamp  time.Time `json:"timestamp"`
	Path       string    `json:"path"`
	RequestID  string    `json:"request_id,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	RequestID  string      `json:"request_id,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}