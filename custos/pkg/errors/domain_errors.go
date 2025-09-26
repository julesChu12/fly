package errors

import "fmt"

const (
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeInvalidPassword    = "INVALID_PASSWORD"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeSessionNotFound    = "SESSION_NOT_FOUND"
	CodeInvalidProvider    = "INVALID_PROVIDER"
)

type DomainError struct {
	Code    string
	Message string
	Fields  map[string]interface{}
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewUserNotFoundError() *DomainError {
	return &DomainError{
		Code:    CodeUserNotFound,
		Message: "User not found",
	}
}

func NewUserAlreadyExistsError(username string) *DomainError {
	return &DomainError{
		Code:    CodeUserAlreadyExists,
		Message: "User already exists",
		Fields:  map[string]interface{}{"username": username},
	}
}

func NewInvalidCredentialsError() *DomainError {
	return &DomainError{
		Code:    CodeInvalidCredentials,
		Message: "Invalid username or password",
	}
}

func NewInvalidPasswordError(reason string) *DomainError {
	return &DomainError{
		Code:    CodeInvalidPassword,
		Message: reason,
	}
}

func NewTokenExpiredError() *DomainError {
	return &DomainError{
		Code:    CodeTokenExpired,
		Message: "Token has expired",
	}
}

func NewTokenInvalidError() *DomainError {
	return &DomainError{
		Code:    CodeTokenInvalid,
		Message: "Token is invalid",
	}
}

func NewSessionNotFoundError() *DomainError {
	return &DomainError{
		Code:    CodeSessionNotFound,
		Message: "Session not found",
	}
}

func NewInvalidProviderError(provider string) *DomainError {
	return &DomainError{
		Code:    CodeInvalidProvider,
		Message: "Invalid OAuth provider",
		Fields:  map[string]interface{}{"provider": provider},
	}
}
