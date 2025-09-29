package password

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Policy represents password policy configuration
type Policy struct {
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	ForbidCommonPasswords bool `json:"forbid_common_passwords"`
	MaxRepeatingChars   int  `json:"max_repeating_chars"`
}

// DefaultPolicy returns the standard security policy
func DefaultPolicy() Policy {
	return Policy{
		MinLength:           8,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		ForbidCommonPasswords: true,
		MaxRepeatingChars:   3,
	}
}

// ValidationResult represents the result of password validation
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
	Score   int      `json:"score"` // 0-100 strength score
}

// Validator handles password validation
type Validator struct {
	policy Policy
	commonPasswords map[string]bool
}

// NewValidator creates a new password validator with the given policy
func NewValidator(policy Policy) *Validator {
	return &Validator{
		policy: policy,
		commonPasswords: getCommonPasswords(),
	}
}

// Validate validates a password against the configured policy
func (v *Validator) Validate(password string) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  make([]string, 0),
		Score:   0,
	}

	// Check minimum length
	if len(password) < v.policy.MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters long", v.policy.MinLength))
	} else {
		result.Score += 20 // Base score for meeting minimum length
	}

	// Check uppercase requirement
	if v.policy.RequireUppercase && !hasUppercase(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	} else if hasUppercase(password) {
		result.Score += 15
	}

	// Check lowercase requirement
	if v.policy.RequireLowercase && !hasLowercase(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	} else if hasLowercase(password) {
		result.Score += 15
	}

	// Check numbers requirement
	if v.policy.RequireNumbers && !hasNumbers(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
	} else if hasNumbers(password) {
		result.Score += 15
	}

	// Check special characters requirement
	if v.policy.RequireSpecialChars && !hasSpecialChars(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character (!@#$%^&*)")
	} else if hasSpecialChars(password) {
		result.Score += 15
	}

	// Check for common passwords
	if v.policy.ForbidCommonPasswords && v.isCommonPassword(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password is too common, please choose a more unique password")
	}

	// Check for excessive repeating characters
	if v.policy.MaxRepeatingChars > 0 && hasExcessiveRepeating(password, v.policy.MaxRepeatingChars) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password cannot have more than %d consecutive repeating characters", v.policy.MaxRepeatingChars))
	}

	// Additional scoring based on length and complexity
	if len(password) >= 12 {
		result.Score += 10 // Bonus for longer passwords
	}
	if len(password) >= 16 {
		result.Score += 10 // Additional bonus for very long passwords
	}

	// Bonus for mixed character types
	charTypes := 0
	if hasUppercase(password) { charTypes++ }
	if hasLowercase(password) { charTypes++ }
	if hasNumbers(password) { charTypes++ }
	if hasSpecialChars(password) { charTypes++ }

	result.Score += charTypes * 5 // Bonus for character diversity

	// Cap the score at 100
	if result.Score > 100 {
		result.Score = 100
	}

	return result
}

// GetPolicy returns the current password policy
func (v *Validator) GetPolicy() Policy {
	return v.policy
}

// UpdatePolicy updates the password policy
func (v *Validator) UpdatePolicy(policy Policy) {
	v.policy = policy
}

// Helper functions

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasNumbers(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecialChars(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}

func hasExcessiveRepeating(s string, maxRepeating int) bool {
	if len(s) < maxRepeating+1 {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count > maxRepeating {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

func (v *Validator) isCommonPassword(password string) bool {
	// Check exact match and common variations
	lower := strings.ToLower(password)
	return v.commonPasswords[lower] ||
		   v.commonPasswords[password] ||
		   isSequentialOrRepeating(password)
}

func isSequentialOrRepeating(password string) bool {
	// Check for simple patterns
	patterns := []string{
		"123456", "abcdef", "qwerty", "asdfgh",
		"111111", "aaaaaa", "000000",
	}

	lower := strings.ToLower(password)
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Check for simple sequences
	if matched, _ := regexp.MatchString(`^(.)\1{5,}$`, password); matched {
		return true
	}

	return false
}

// getCommonPasswords returns a set of common passwords to forbid
func getCommonPasswords() map[string]bool {
	// Top 100+ most common passwords
	passwords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "1234567890", "abc123",
		"111111", "dragon", "master", "princess", "login",
		"passw0rd", "password1", "123456789", "welcome123", "admin123",
		"guest", "test", "user", "root", "toor", "pass",
		"12345", "1234", "123", "password12", "qwerty123",
		"football", "baseball", "basketball", "soccer", "tennis",
		"iloveyou", "sunshine", "superman", "batman", "spiderman",
		"computer", "internet", "nintendo", "pokemon", "mustang",
		"jordan", "hunter", "ranger", "shadow", "tiger",
		"trustno1", "starwars", "freedom", "whatever", "secret",
		"summer", "winter", "spring", "autumn", "hello",
		"world", "earth", "fire", "water", "nature",
		"family", "friend", "love", "peace", "happy",
		"lucky", "money", "golden", "silver", "diamond",
		"flower", "garden", "mountain", "ocean", "river",
		"default", "changeme", "newpass", "temp", "temporary",
		"sample", "demo", "example", "public", "private",
	}

	result := make(map[string]bool)
	for _, pwd := range passwords {
		result[pwd] = true
		result[strings.ToUpper(pwd)] = true
		result[strings.Title(pwd)] = true
	}

	return result
}

// GenerateStrengthMessage returns a human-readable strength message
func GenerateStrengthMessage(score int) string {
	switch {
	case score >= 90:
		return "Very Strong"
	case score >= 75:
		return "Strong"
	case score >= 60:
		return "Good"
	case score >= 40:
		return "Fair"
	case score >= 25:
		return "Weak"
	default:
		return "Very Weak"
	}
}