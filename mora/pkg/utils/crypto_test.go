package utils

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 8", 8},
		{"length 16", 16},
		{"length 32", 32},
		{"length 0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Errorf("GenerateRandomString() error = %v", err)
				return
			}

			if len(result) != tt.length {
				t.Errorf("GenerateRandomString() length = %v, want %v", len(result), tt.length)
			}

			// Test uniqueness by generating multiple strings
			if tt.length > 0 {
				result2, err := GenerateRandomString(tt.length)
				if err != nil {
					t.Errorf("GenerateRandomString() second call error = %v", err)
					return
				}

				if result == result2 && tt.length > 4 {
					t.Error("GenerateRandomString() should generate unique strings")
				}
			}
		})
	}
}

func TestGenerateTraceID(t *testing.T) {
	traceID := GenerateTraceID()

	if !strings.HasPrefix(traceID, "trace-") {
		t.Errorf("GenerateTraceID() = %v, should start with 'trace-'", traceID)
	}

	// Test uniqueness
	traceID2 := GenerateTraceID()
	if traceID == traceID2 {
		t.Error("GenerateTraceID() should generate unique trace IDs")
	}
}

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:  "hello world",
			input: "hello world",
			want:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:  "test string",
			input: "test",
			want:  "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashSHA256(tt.input); got != tt.want {
				t.Errorf("HashSHA256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short string",
			input: "123",
			want:  "***",
		},
		{
			name:  "8 char string",
			input: "12345678",
			want:  "********",
		},
		{
			name:  "long string",
			input: "1234567890abcdef",
			want:  "1234********cdef",
		},
		{
			name:  "very long string",
			input: "1234567890abcdefghijklmnop",
			want:  "1234******************mnop",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskSensitive(tt.input); got != tt.want {
				t.Errorf("MaskSensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}
