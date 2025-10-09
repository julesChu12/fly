package errors

import "errors"

// NotFoundError 表示资源未找到的错误
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrContactNotFound  = errors.New("contact not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInternalServer   = errors.New("internal server error")
)