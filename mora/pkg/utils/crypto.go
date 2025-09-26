package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateTraceID generates a unique trace ID for request tracking
func GenerateTraceID() string {
	id, _ := GenerateRandomString(16)
	return fmt.Sprintf("trace-%d-%s", time.Now().UnixNano(), id)
}

// HashSHA256 creates a SHA256 hash of the input string
func HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// MaskSensitive masks sensitive information for logging
func MaskSensitive(input string) string {
	if len(input) <= 8 {
		return strings.Repeat("*", len(input))
	}
	return input[:4] + strings.Repeat("*", len(input)-8) + input[len(input)-4:]
}
