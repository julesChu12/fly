package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// HashConfig represents the configuration for password hashing
type HashConfig struct {
	Memory      uint32 `json:"memory"`      // Memory usage in KB
	Iterations  uint32 `json:"iterations"`  // Number of iterations
	Parallelism uint8  `json:"parallelism"` // Number of threads
	SaltLength  uint32 `json:"salt_length"` // Salt length in bytes
	KeyLength   uint32 `json:"key_length"`  // Key length in bytes
}

// DefaultHashConfig returns the recommended configuration for Argon2id
func DefaultHashConfig() HashConfig {
	return HashConfig{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,         // 3 iterations
		Parallelism: 2,         // 2 threads
		SaltLength:  16,        // 16 bytes salt
		KeyLength:   32,        // 32 bytes key
	}
}

// Hasher handles password hashing and verification
type Hasher struct {
	config HashConfig
}

// NewHasher creates a new password hasher with the given configuration
func NewHasher(config HashConfig) *Hasher {
	return &Hasher{
		config: config,
	}
}

// Hash generates a secure hash of the given password using Argon2id
func (h *Hasher) Hash(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Iterations,
		h.config.Memory,
		h.config.Parallelism,
		h.config.KeyLength,
	)

	// Encode the result
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	hashedPassword := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.config.Memory,
		h.config.Iterations,
		h.config.Parallelism,
		encodedSalt,
		encodedHash,
	)

	return hashedPassword, nil
}

// Verify checks if the given password matches the stored hash
func (h *Hasher) Verify(password, hashedPassword string) (bool, error) {
	// Parse the hashed password
	config, salt, hash, err := h.parseHash(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Generate hash with the same parameters
	testHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, testHash) == 1, nil
}

// parseHash parses a hashed password string and extracts configuration, salt, and hash
func (h *Hasher) parseHash(hashedPassword string) (HashConfig, []byte, []byte, error) {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return HashConfig{}, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return HashConfig{}, nil, nil, fmt.Errorf("unsupported hash algorithm: %s", parts[1])
	}

	if parts[2] != "v=19" {
		return HashConfig{}, nil, nil, fmt.Errorf("unsupported argon2 version: %s", parts[2])
	}

	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return HashConfig{}, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return HashConfig{}, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return HashConfig{}, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	config := HashConfig{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  uint32(len(salt)),
		KeyLength:   uint32(len(hash)),
	}

	return config, salt, hash, nil
}

// IsHashValid checks if a hash string is in the correct format
func (h *Hasher) IsHashValid(hashedPassword string) bool {
	_, _, _, err := h.parseHash(hashedPassword)
	return err == nil
}

// GetConfig returns the current hashing configuration
func (h *Hasher) GetConfig() HashConfig {
	return h.config
}

// UpdateConfig updates the hashing configuration
func (h *Hasher) UpdateConfig(config HashConfig) {
	h.config = config
}

// PasswordService combines validation and hashing functionality
type PasswordService struct {
	validator *Validator
	hasher    *Hasher
}

// NewPasswordService creates a new password service with default configurations
func NewPasswordService() *PasswordService {
	return &PasswordService{
		validator: NewValidator(DefaultPolicy()),
		hasher:    NewHasher(DefaultHashConfig()),
	}
}

// NewPasswordServiceWithConfig creates a new password service with custom configurations
func NewPasswordServiceWithConfig(policy Policy, hashConfig HashConfig) *PasswordService {
	return &PasswordService{
		validator: NewValidator(policy),
		hasher:    NewHasher(hashConfig),
	}
}

// ValidateAndHash validates a password and hashes it if valid
func (ps *PasswordService) ValidateAndHash(password string) (string, ValidationResult, error) {
	// Validate password
	result := ps.validator.Validate(password)
	if !result.IsValid {
		return "", result, nil
	}

	// Hash password
	hashedPassword, err := ps.hasher.Hash(password)
	if err != nil {
		return "", result, fmt.Errorf("failed to hash password: %w", err)
	}

	return hashedPassword, result, nil
}

// VerifyPassword verifies a password against its hash
func (ps *PasswordService) VerifyPassword(password, hashedPassword string) (bool, error) {
	return ps.hasher.Verify(password, hashedPassword)
}

// ValidatePassword validates a password without hashing
func (ps *PasswordService) ValidatePassword(password string) ValidationResult {
	return ps.validator.Validate(password)
}

// GetValidator returns the password validator
func (ps *PasswordService) GetValidator() *Validator {
	return ps.validator
}

// GetHasher returns the password hasher
func (ps *PasswordService) GetHasher() *Hasher {
	return ps.hasher
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	UserID          uint      `json:"user_id"`
	CurrentPassword string    `json:"current_password"`
	NewPassword     string    `json:"new_password"`
	Timestamp       time.Time `json:"timestamp"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email       string    `json:"email"`
	Token       string    `json:"token"`
	NewPassword string    `json:"new_password"`
	ExpiresAt   time.Time `json:"expires_at"`
}
