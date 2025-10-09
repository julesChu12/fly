package errors

import "errors"

// Custom error types
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type DuplicateError struct {
	Message string
}

func (e *DuplicateError) Error() string {
	return e.Message
}

type BusinessError struct {
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// Predefined errors
var (
	// Order errors
	ErrOrderNotFound         = &NotFoundError{Message: "order not found"}
	ErrOrderItemNotFound     = &NotFoundError{Message: "order item not found"}
	ErrDuplicateOrderNo      = &DuplicateError{Message: "order number already exists"}
	ErrInvalidOrderStatus    = &ValidationError{Message: "invalid order status"}
	ErrOrderCannotBeModified = &BusinessError{Message: "order cannot be modified in current status"}
	ErrInvalidStatusTransition = &BusinessError{Message: "invalid status transition"}
	ErrEmptyOrderItems       = &ValidationError{Message: "order must have at least one item"}
	ErrInvalidAmount         = &ValidationError{Message: "invalid amount"}
	ErrInvalidQuantity       = &ValidationError{Message: "invalid quantity"}

	// General errors
	ErrInvalidRequest = &ValidationError{Message: "invalid request"}
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternalError  = errors.New("internal server error")
)

// Error codes for API responses
const (
	CodeSuccess      = 0
	CodeInvalidParam = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeDuplicate    = 409
	CodeBusiness     = 422
	CodeInternal     = 500
)

// GetErrorCode returns appropriate HTTP status code for error
func GetErrorCode(err error) int {
	switch err.(type) {
	case *NotFoundError:
		return CodeNotFound
	case *ValidationError:
		return CodeInvalidParam
	case *DuplicateError:
		return CodeDuplicate
	case *BusinessError:
		return CodeBusiness
	default:
		if err == ErrUnauthorized {
			return CodeUnauthorized
		}
		if err == ErrForbidden {
			return CodeForbidden
		}
		return CodeInternal
	}
}