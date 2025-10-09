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

type InsufficientBalanceError struct {
	Message string
	Current float64
	Required float64
}

func (e *InsufficientBalanceError) Error() string {
	return e.Message
}

// Predefined errors
var (
	// Wallet errors
	ErrWalletNotFound       = &NotFoundError{Message: "wallet not found"}
	ErrWalletFrozen         = &BusinessError{Message: "wallet is frozen"}
	ErrWalletAlreadyExists  = &DuplicateError{Message: "wallet already exists for this customer"}
	ErrInsufficientBalance  = &InsufficientBalanceError{Message: "insufficient balance"}

	// Transaction errors
	ErrTransactionNotFound    = &NotFoundError{Message: "transaction not found"}
	ErrDuplicateTransaction   = &DuplicateError{Message: "transaction with this idempotency key already exists"}
	ErrInvalidAmount         = &ValidationError{Message: "invalid amount"}
	ErrInvalidTransactionType = &ValidationError{Message: "invalid transaction type"}
	ErrTransactionFailed      = &BusinessError{Message: "transaction failed"}

	// Payment channel errors
	ErrPaymentChannelNotFound = &NotFoundError{Message: "payment channel not found"}
	ErrChannelDisabled        = &BusinessError{Message: "payment channel is disabled"}

	// General errors
	ErrInvalidRequest = &ValidationError{Message: "invalid request"}
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternalError  = errors.New("internal server error")
)

// Error codes for API responses
const (
	CodeSuccess            = 0
	CodeInvalidParam       = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeDuplicate          = 409
	CodeBusiness           = 422
	CodeInsufficientFunds  = 423
	CodeInternal           = 500
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
	case *InsufficientBalanceError:
		return CodeInsufficientFunds
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