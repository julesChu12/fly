package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

// ProfileUpdateRequest represents validation rules for profile updates
type ProfileUpdateRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
	Bio       string `json:"bio"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Website   string `json:"website"`
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// ValidateProfileUpdate validates profile update request data
func ValidateProfileUpdate(req ProfileUpdateRequest) ValidationErrors {
	var errors ValidationErrors

	// Validate username
	if req.Username != "" {
		if err := validateUsername(req.Username); err != nil {
			errors = append(errors, ValidationError{
				Field:   "username",
				Message: err.Error(),
				Code:    "invalid_username",
			})
		}
	}

	// Validate email
	if req.Email != "" {
		if err := validateEmail(req.Email); err != nil {
			errors = append(errors, ValidationError{
				Field:   "email",
				Message: err.Error(),
				Code:    "invalid_email",
			})
		}
	}

	// Validate first name
	if req.FirstName != "" {
		if err := validateName(req.FirstName, "first name"); err != nil {
			errors = append(errors, ValidationError{
				Field:   "first_name",
				Message: err.Error(),
				Code:    "invalid_first_name",
			})
		}
	}

	// Validate last name
	if req.LastName != "" {
		if err := validateName(req.LastName, "last name"); err != nil {
			errors = append(errors, ValidationError{
				Field:   "last_name",
				Message: err.Error(),
				Code:    "invalid_last_name",
			})
		}
	}

	// Validate avatar URL
	if req.Avatar != "" {
		if err := validateURL(req.Avatar, "avatar"); err != nil {
			errors = append(errors, ValidationError{
				Field:   "avatar",
				Message: err.Error(),
				Code:    "invalid_avatar_url",
			})
		}
	}

	// Validate bio
	if req.Bio != "" {
		if err := validateBio(req.Bio); err != nil {
			errors = append(errors, ValidationError{
				Field:   "bio",
				Message: err.Error(),
				Code:    "invalid_bio",
			})
		}
	}

	// Validate phone
	if req.Phone != "" {
		if err := validatePhone(req.Phone); err != nil {
			errors = append(errors, ValidationError{
				Field:   "phone",
				Message: err.Error(),
				Code:    "invalid_phone",
			})
		}
	}

	// Validate location
	if req.Location != "" {
		if err := validateLocation(req.Location); err != nil {
			errors = append(errors, ValidationError{
				Field:   "location",
				Message: err.Error(),
				Code:    "invalid_location",
			})
		}
	}

	// Validate website URL
	if req.Website != "" {
		if err := validateURL(req.Website, "website"); err != nil {
			errors = append(errors, ValidationError{
				Field:   "website",
				Message: err.Error(),
				Code:    "invalid_website_url",
			})
		}
	}

	return errors
}

// ValidatePreferences validates user preferences
func ValidatePreferences(preferences map[string]string) ValidationErrors {
	var errors ValidationErrors

	// Define allowed preference keys and their validation rules
	allowedKeys := map[string]func(string) error{
		"language":      validateLanguage,
		"timezone":      validateTimezone,
		"theme":         validateTheme,
		"notifications": validateNotificationSetting,
	}

	for key, value := range preferences {
		// Check if key is allowed
		validator, allowed := allowedKeys[key]
		if !allowed {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: fmt.Sprintf("Unknown preference key: %s", key),
				Code:    "unknown_preference_key",
			})
			continue
		}

		// Validate value
		if err := validator(value); err != nil {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: err.Error(),
				Code:    "invalid_preference_value",
			})
		}
	}

	return errors
}

// Individual validation functions

func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 30 {
		return fmt.Errorf("username must be no more than 30 characters long")
	}

	// Username can contain letters, numbers, underscores, and hyphens
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}

func validateEmail(email string) error {
	if len(email) > 254 {
		return fmt.Errorf("email address is too long")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address format")
	}

	return nil
}

func validateName(name, fieldName string) error {
	if len(name) > 50 {
		return fmt.Errorf("%s must be no more than 50 characters long", fieldName)
	}

	// Names can contain letters, spaces, hyphens, and apostrophes
	matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, name)
	if !matched {
		return fmt.Errorf("%s can only contain letters, spaces, hyphens, and apostrophes", fieldName)
	}

	return nil
}

func validateURL(url, fieldName string) error {
	if len(url) > 512 {
		return fmt.Errorf("%s URL is too long", fieldName)
	}

	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
	if !matched {
		return fmt.Errorf("invalid %s URL format", fieldName)
	}

	return nil
}

func validateBio(bio string) error {
	if len(bio) > 500 {
		return fmt.Errorf("bio must be no more than 500 characters long")
	}

	return nil
}

func validatePhone(phone string) error {
	if len(phone) > 20 {
		return fmt.Errorf("phone number is too long")
	}

	// Basic phone validation (international format)
	matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, phone)
	if !matched {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

func validateLocation(location string) error {
	if len(location) > 100 {
		return fmt.Errorf("location must be no more than 100 characters long")
	}

	return nil
}

func validateLanguage(language string) error {
	allowedLanguages := []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"}
	for _, allowed := range allowedLanguages {
		if language == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported language: %s", language)
}

func validateTimezone(timezone string) error {
	// Basic timezone validation - in a real implementation, you'd check against a list of valid timezones
	if len(timezone) == 0 || len(timezone) > 50 {
		return fmt.Errorf("invalid timezone format")
	}
	return nil
}

func validateTheme(theme string) error {
	allowedThemes := []string{"light", "dark", "auto"}
	for _, allowed := range allowedThemes {
		if theme == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported theme: %s", theme)
}

func validateNotificationSetting(setting string) error {
	allowedSettings := []string{"enabled", "disabled", "email_only", "push_only"}
	for _, allowed := range allowedSettings {
		if setting == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported notification setting: %s", setting)
}