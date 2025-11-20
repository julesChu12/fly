package errors

import (
	"fmt"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// TestNotFoundError_Error tests the Error() method of NotFoundError
func TestNotFoundError_Error(t *testing.T) {
	// Arrange
	message := "customer with ID 123 not found"
	notFoundErr := &NotFoundError{
		Message: message,
	}

	// Act
	result := notFoundErr.Error()

	// Assert
	assert.Equal(t, message, result)
}

// TestNotFoundError_Creation tests creating a NotFoundError
func TestNotFoundError_Creation(t *testing.T) {
	// Arrange & Act
	message := "test resource not found"
	notFoundErr := &NotFoundError{
		Message: message,
	}

	// Assert
	assert.Equal(t, message, notFoundErr.Message)
	assert.Equal(t, message, notFoundErr.Error())
}

// TestPredefinedErrors tests all predefined error variables
func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name:  "ErrCustomerNotFound",
			error: ErrCustomerNotFound,
		},
		{
			name:  "ErrContactNotFound",
			error: ErrContactNotFound,
		},
		{
			name:  "ErrInvalidInput",
			error: ErrInvalidInput,
		},
		{
			name:  "ErrDuplicateEmail",
			error: ErrDuplicateEmail,
		},
		{
			name:  "ErrInternalServer",
			error: ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert
			assert.NotNil(t, tt.error)
			assert.NotEmpty(t, tt.error.Error())
		})
	}
}

// TestErrorMessages verifies error messages are descriptive
func TestErrorMessages(t *testing.T) {
	// Test predefined error messages
	assert.Equal(t, "customer not found", ErrCustomerNotFound.Error())
	assert.Equal(t, "contact not found", ErrContactNotFound.Error())
	assert.Equal(t, "invalid input", ErrInvalidInput.Error())
	assert.Equal(t, "email already exists", ErrDuplicateEmail.Error())
	assert.Equal(t, "internal server error", ErrInternalServer.Error())
}

// TestErrorTypeChecking tests error type checking and assertions
func TestErrorTypeChecking(t *testing.T) {
	// Test NotFoundError type checking
	customErr := &NotFoundError{Message: "custom not found error"}

	// Should satisfy error interface
	var _ error = customErr
	assert.Implements(t, (*error)(nil), customErr)

	// Test error wrapping and unwrapping
	wrappedErr := pkgerrors.Wrap(customErr, "wrapped error")
	assert.NotNil(t, wrappedErr)

	// Test that predefined errors are not nil
	assert.NotNil(t, ErrCustomerNotFound)
	assert.NotNil(t, ErrContactNotFound)
	assert.NotNil(t, ErrInvalidInput)
	assert.NotNil(t, ErrDuplicateEmail)
	assert.NotNil(t, ErrInternalServer)
}

// TestErrorConsistency tests that errors are consistent across calls
func TestErrorConsistency(t *testing.T) {
	// Test that multiple references to the same error are identical
	assert.Equal(t, ErrCustomerNotFound, ErrCustomerNotFound)
	assert.Equal(t, ErrContactNotFound, ErrContactNotFound)
	assert.Equal(t, ErrInvalidInput, ErrInvalidInput)
	assert.Equal(t, ErrDuplicateEmail, ErrDuplicateEmail)
	assert.Equal(t, ErrInternalServer, ErrInternalServer)
}

// TestCustomNotFoundErrorCreation tests creating custom not found errors
func TestCustomNotFoundErrorCreation(t *testing.T) {
	// Test creating various custom not found errors
	testCases := []struct {
		resourceType string
		resourceID   string
	}{
		{"customer", "123"},
		{"contact", "456"},
		{"user", "789"},
		{"order", "101112"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_not_found_%s", tc.resourceType, tc.resourceID), func(t *testing.T) {
			message := fmt.Sprintf("%s with ID %s not found", tc.resourceType, tc.resourceID)
			customErr := &NotFoundError{Message: message}

			assert.Equal(t, message, customErr.Error())
			assert.Contains(t, customErr.Error(), tc.resourceType)
			assert.Contains(t, customErr.Error(), tc.resourceID)
		})
	}
}

// TestErrorHandlingScenario tests realistic error handling scenarios
func TestErrorHandlingScenario(t *testing.T) {
	// Simulate a service function that might return different errors
	findCustomer := func(id uint) error {
		if id == 0 {
			return ErrInvalidInput
		}
		if id == 999 {
			return ErrCustomerNotFound
		}
		return nil
	}

	// Test different scenarios
	scenarios := []struct {
		name         string
		customerID   uint
		expectedErr  error
		shouldError  bool
	}{
		{
			name:        "valid customer",
			customerID:  123,
			expectedErr: nil,
			shouldError: false,
		},
		{
			name:        "invalid ID",
			customerID:  0,
			expectedErr: ErrInvalidInput,
			shouldError: true,
		},
		{
			name:        "customer not found",
			customerID:  999,
			expectedErr: ErrCustomerNotFound,
			shouldError: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := findCustomer(scenario.customerID)

			if scenario.shouldError {
				assert.Error(t, err)
				assert.Equal(t, scenario.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkNotFoundError_Error benchmarks the Error() method
func BenchmarkNotFoundError_Error(b *testing.B) {
	err := &NotFoundError{Message: "benchmark test error"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// BenchmarkPredefinedErrors benchmarks accessing predefined errors
func BenchmarkPredefinedErrors(b *testing.B) {
	errors := []error{
		ErrCustomerNotFound,
		ErrContactNotFound,
		ErrInvalidInput,
		ErrDuplicateEmail,
		ErrInternalServer,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors[i%len(errors)].Error()
	}
}