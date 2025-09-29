package errors

import "errors"

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrContactNotFound  = errors.New("contact not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInternalServer   = errors.New("internal server error")
)