package utils

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tabs and spaces", "\t \n ", true},
		{"normal string", "hello", false},
		{"string with spaces", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.input); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"snake_case", "hello_world", "helloWorld"},
		{"multiple underscores", "hello_world_test", "helloWorldTest"},
		{"leading underscore", "_hello_world", "HelloWorld"},
		{"trailing underscore", "hello_world_", "helloWorld"},
		{"already camelCase", "helloWorld", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamelCase(tt.input); got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"camelCase", "helloWorld", "hello_world"},
		{"PascalCase", "HelloWorld", "hello_world"},
		{"multiple words", "HelloWorldTest", "hello_world_test"},
		{"already snake_case", "hello_world", "hello_world"},
		{"with numbers", "Hello2World", "hello2_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.input); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short limit", "hello", 3, "..."},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLength); got != tt.want {
				t.Errorf("Truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name string
		item string
		want bool
	}{
		{"exists", "banana", true},
		{"not exists", "grape", false},
		{"empty item", "", false},
		{"case sensitive", "Apple", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(slice, tt.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with empty slice
	emptySlice := []string{}
	if Contains(emptySlice, "test") {
		t.Error("Contains() should return false for empty slice")
	}
}
