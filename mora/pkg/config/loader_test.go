package config

import (
	"os"
	"reflect"
	"testing"
)

type TestConfig struct {
	Database struct {
		Host     string `yaml:"host" env:"DB_HOST"`
		Port     int    `yaml:"port" env:"DB_PORT"`
		Username string `yaml:"username" env:"DB_USERNAME"`
		Password string `yaml:"password" env:"DB_PASSWORD"`
	} `yaml:"database"`
	Server struct {
		Port    int  `yaml:"port" env:"SERVER_PORT"`
		Debug   bool `yaml:"debug" env:"DEBUG"`
		Timeout int  `yaml:"timeout" env:"TIMEOUT"`
	} `yaml:"server"`
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"DB_HOST":      "localhost",
		"DB_PORT":      "5432",
		"DB_USERNAME":  "testuser",
		"DB_PASSWORD":  "testpass",
		"SERVER_PORT":  "8080",
		"DEBUG":        "true",
		"TIMEOUT":      "30",
	}

	// Set env vars
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	var cfg TestConfig
	loader := NewLoader()
	err := loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify values
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want 5432", cfg.Database.Port)
	}
	if cfg.Database.Username != "testuser" {
		t.Errorf("Database.Username = %v, want testuser", cfg.Database.Username)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Debug != true {
		t.Errorf("Server.Debug = %v, want true", cfg.Server.Debug)
	}
}

func TestLoadConfig_FromYAML(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
database:
  host: yaml-host
  port: 3306
  username: yaml-user
  password: yaml-pass
server:
  port: 9000
  debug: false
  timeout: 60
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	var cfg TestConfig
	loader := NewLoader(WithConfigPaths(tmpFile.Name()))
	err = loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify values
	if cfg.Database.Host != "yaml-host" {
		t.Errorf("Database.Host = %v, want yaml-host", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("Database.Port = %v, want 3306", cfg.Database.Port)
	}
	if cfg.Server.Debug != false {
		t.Errorf("Server.Debug = %v, want false", cfg.Server.Debug)
	}
}

func TestLoadConfig_EnvOverridesYAML(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
database:
  host: yaml-host
  port: 3306
server:
  port: 9000
  debug: false
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Set some environment variables that should override YAML
	os.Setenv("DB_HOST", "env-host")
	os.Setenv("SERVER_PORT", "8080")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("SERVER_PORT")
	}()

	var cfg TestConfig
	loader := NewLoader(WithConfigPaths(tmpFile.Name()))
	err = loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Environment variables should override YAML values
	if cfg.Database.Host != "env-host" {
		t.Errorf("Database.Host = %v, want env-host", cfg.Database.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", cfg.Server.Port)
	}
	// YAML values should remain for fields not overridden by env
	if cfg.Database.Port != 3306 {
		t.Errorf("Database.Port = %v, want 3306", cfg.Database.Port)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "env var exists",
			key:          "TEST_KEY_EXISTS",
			defaultValue: "default",
			envValue:     "exists",
			want:         "exists",
		},
		{
			name:         "env var not exists",
			key:          "TEST_KEY_NOT_EXISTS",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			if got := GetEnv(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvWithPrefix(t *testing.T) {
	os.Setenv("MYAPP_TEST_KEY", "prefixed_value")
	defer os.Unsetenv("MYAPP_TEST_KEY")

	got := GetEnvWithPrefix("MYAPP", "TEST_KEY", "default")
	if got != "prefixed_value" {
		t.Errorf("GetEnvWithPrefix() = %v, want prefixed_value", got)
	}

	// Test with non-existing key
	got = GetEnvWithPrefix("MYAPP", "NOT_EXISTS", "default")
	if got != "default" {
		t.Errorf("GetEnvWithPrefix() = %v, want default", got)
	}
}

func TestWithConfigPaths(t *testing.T) {
	paths := []string{"path1", "path2", "path3"}
	loader := NewLoader(WithConfigPaths(paths...))

	if !reflect.DeepEqual(loader.configPaths, paths) {
		t.Errorf("WithConfigPaths() = %v, want %v", loader.configPaths, paths)
	}
}

func TestWithEnvPrefix(t *testing.T) {
	prefix := "MYAPP"
	loader := NewLoader(WithEnvPrefix(prefix))

	if loader.envPrefix != prefix {
		t.Errorf("WithEnvPrefix() = %v, want %v", loader.envPrefix, prefix)
	}
}

func TestMustLoadConfig_Success(t *testing.T) {
	os.Setenv("SERVER_PORT", "3000")
	defer os.Unsetenv("SERVER_PORT")

	var cfg TestConfig

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustLoadConfig() panicked: %v", r)
		}
	}()

	MustLoadConfig(&cfg)

	if cfg.Server.Port != 3000 {
		t.Errorf("MustLoadConfig() Server.Port = %v, want 3000", cfg.Server.Port)
	}
}

func TestSetFieldValue(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name      string
		fieldType reflect.Kind
		value     string
		wantError bool
	}{
		{"string", reflect.String, "test", false},
		{"int", reflect.Int, "123", false},
		{"bool true", reflect.Bool, "true", false},
		{"bool false", reflect.Bool, "false", false},
		{"float", reflect.Float64, "12.34", false},
		{"invalid int", reflect.Int, "invalid", true},
		{"invalid bool", reflect.Bool, "invalid", true},
		{"invalid float", reflect.Float64, "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reflect.Value of the appropriate type
			var field reflect.Value
			switch tt.fieldType {
			case reflect.String:
				var s string
				field = reflect.ValueOf(&s).Elem()
			case reflect.Int:
				var i int
				field = reflect.ValueOf(&i).Elem()
			case reflect.Bool:
				var b bool
				field = reflect.ValueOf(&b).Elem()
			case reflect.Float64:
				var f float64
				field = reflect.ValueOf(&f).Elem()
			}

			err := loader.setFieldValue(field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("setFieldValue() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}