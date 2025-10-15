package entity

import "errors"

// Domain entity errors
var (
	// ErrInvalidGender is returned when an invalid gender value is provided
	ErrInvalidGender = errors.New("invalid gender value, must be one of: male, female, other")

	// ErrInvalidEmail is returned when an invalid email format is provided
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidPhone is returned when an invalid phone format is provided
	ErrInvalidPhone = errors.New("invalid phone format")

	// ErrWeakPassword is returned when a password doesn't meet security requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters long")
)
