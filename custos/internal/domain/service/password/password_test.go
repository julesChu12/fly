package password

import (
	"strings"
	"testing"
)

func TestPasswordValidator(t *testing.T) {
	validator := NewValidator(DefaultPolicy())

	tests := []struct {
		name     string
		password string
		valid    bool
		errors   int
	}{
		{
			name:     "Strong password",
			password: "MySecure#Pass123!",
			valid:    true,
			errors:   0,
		},
		{
			name:     "Too short",
			password: "Abc1!",
			valid:    false,
			errors:   1,
		},
		{
			name:     "No uppercase",
			password: "mysecure#pass123!",
			valid:    false,
			errors:   1,
		},
		{
			name:     "No lowercase",
			password: "MYSECURE#PASS123!",
			valid:    false,
			errors:   1,
		},
		{
			name:     "No numbers",
			password: "MySecure#Pass!",
			valid:    false,
			errors:   1,
		},
		{
			name:     "No special chars",
			password: "MySecurePass123",
			valid:    false,
			errors:   1,
		},
		{
			name:     "Common password",
			password: "password123",
			valid:    false,
			errors:   3, // no uppercase, no special chars, common password
		},
		{
			name:     "Excessive repeating",
			password: "MySecuuuure#Pass123!",
			valid:    false,
			errors:   1,
		},
		{
			name:     "Multiple errors",
			password: "pass",
			valid:    false,
			errors:   5, // short, no uppercase, no numbers, no special chars, common password
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.password)

			if result.IsValid != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, result.IsValid)
			}

			if len(result.Errors) != tt.errors {
				t.Errorf("Expected %d errors, got %d errors: %v", tt.errors, len(result.Errors), result.Errors)
			}

			if tt.valid && result.Score < 60 {
				t.Errorf("Valid password should have score >= 60, got %d", result.Score)
			}
		})
	}
}

func TestPasswordHasher(t *testing.T) {
	hasher := NewHasher(DefaultHashConfig())
	password := "MySecure#Pass123!"

	// Test hashing
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("Hash should start with argon2id identifier, got: %s", hash[:20])
	}

	// Test verification with correct password
	valid, err := hasher.Verify(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !valid {
		t.Error("Password verification should succeed with correct password")
	}

	// Test verification with incorrect password
	valid, err = hasher.Verify("WrongPassword!", hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Password verification should fail with incorrect password")
	}

	// Test hash validation
	if !hasher.IsHashValid(hash) {
		t.Error("Generated hash should be valid")
	}

	if hasher.IsHashValid("invalid_hash") {
		t.Error("Invalid hash should not be valid")
	}
}

func TestPasswordService(t *testing.T) {
	service := NewPasswordService()

	tests := []struct {
		name     string
		password string
		expectValid bool
	}{
		{
			name:     "Valid password",
			password: "MySecure#Pass123!",
			expectValid: true,
		},
		{
			name:     "Invalid password",
			password: "weak",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, result, err := service.ValidateAndHash(tt.password)

			if tt.expectValid {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !result.IsValid {
					t.Errorf("Expected valid password, got errors: %v", result.Errors)
				}
				if hash == "" {
					t.Error("Expected hash to be generated for valid password")
				}

				// Test verification
				valid, err := service.VerifyPassword(tt.password, hash)
				if err != nil {
					t.Fatalf("Failed to verify password: %v", err)
				}
				if !valid {
					t.Error("Password verification should succeed")
				}
			} else {
				if result.IsValid {
					t.Error("Expected invalid password")
				}
				if hash != "" {
					t.Error("Hash should not be generated for invalid password")
				}
			}
		})
	}
}

func TestGenerateStrengthMessage(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{95, "Very Strong"},
		{80, "Strong"},
		{65, "Good"},
		{50, "Fair"},
		{30, "Weak"},
		{10, "Very Weak"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GenerateStrengthMessage(tt.score)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCustomPolicy(t *testing.T) {
	// Test with relaxed policy
	relaxedPolicy := Policy{
		MinLength:           6,
		RequireUppercase:    false,
		RequireLowercase:    true,
		RequireNumbers:      false,
		RequireSpecialChars: false,
		ForbidCommonPasswords: false,
		MaxRepeatingChars:   5,
	}

	validator := NewValidator(relaxedPolicy)

	// This should be valid with relaxed policy
	result := validator.Validate("simple")
	if !result.IsValid {
		t.Errorf("Simple password should be valid with relaxed policy, errors: %v", result.Errors)
	}

	// Test with strict policy
	strictPolicy := Policy{
		MinLength:           12,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		ForbidCommonPasswords: true,
		MaxRepeatingChars:   2,
	}

	validator.UpdatePolicy(strictPolicy)

	// This should be invalid with strict policy
	result = validator.Validate("MySecure#Pass123!")
	if !result.IsValid {
		t.Errorf("Should be valid with strict policy, errors: %v", result.Errors)
	}

	// Test password that should fail strict policy due to length
	result = validator.Validate("Short#1")
	if result.IsValid {
		t.Error("Should be invalid with strict policy due to length requirement")
	}
}

func BenchmarkPasswordHashing(b *testing.B) {
	hasher := NewHasher(DefaultHashConfig())
	password := "MySecure#Pass123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatalf("Failed to hash password: %v", err)
		}
	}
}

func BenchmarkPasswordVerification(b *testing.B) {
	hasher := NewHasher(DefaultHashConfig())
	password := "MySecure#Pass123!"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Verify(password, hash)
		if err != nil {
			b.Fatalf("Failed to verify password: %v", err)
		}
	}
}