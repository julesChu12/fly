package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading from various sources
type Loader struct {
	configPaths []string
	envPrefix   string
}

// Option represents a configuration option
type Option func(*Loader)

// WithConfigPaths sets the configuration file paths to search
func WithConfigPaths(paths ...string) Option {
	return func(l *Loader) {
		l.configPaths = paths
	}
}

// WithEnvPrefix sets the environment variable prefix
func WithEnvPrefix(prefix string) Option {
	return func(l *Loader) {
		l.envPrefix = prefix
	}
}

// NewLoader creates a new configuration loader
func NewLoader(opts ...Option) *Loader {
	loader := &Loader{
		configPaths: []string{"config.yaml", "config.yml", "./config/config.yaml"},
		envPrefix:   "",
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader
}

// Load loads configuration into the provided struct
func (l *Loader) Load(cfg any) error {
	// First, try to load from YAML files
	if err := l.loadFromFile(cfg); err != nil {
		return fmt.Errorf("failed to load config from file: %w", err)
	}

	// Then, override with environment variables
	if err := l.loadFromEnv(cfg); err != nil {
		return fmt.Errorf("failed to load config from env: %w", err)
	}

	return nil
}

// loadFromFile loads configuration from YAML files
func (l *Loader) loadFromFile(cfg any) error {
	var configFile string
	var found bool

	// Find the first existing config file
	for _, path := range l.configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			found = true
			break
		}
	}

	if !found {
		// No config file found, that's okay - we'll rely on env vars or defaults
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables using reflection
func (l *Loader) loadFromEnv(cfg any) error {
	return l.loadStructFromEnv(reflect.ValueOf(cfg).Elem(), "")
}

// loadStructFromEnv recursively loads struct fields from environment variables
func (l *Loader) loadStructFromEnv(v reflect.Value, prefix string) error {
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct {
			// For nested structs, use the struct field name as prefix
			nestedPrefix := fieldType.Name
			if prefix != "" {
				nestedPrefix = prefix + "_" + fieldType.Name
			}
			if err := l.loadStructFromEnv(field, nestedPrefix); err != nil {
				return err
			}
			continue
		}

		// Get environment variable name from env tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue // Skip fields without env tag
		}

		// Get environment variable value
		envValue := os.Getenv(envTag)
		if envValue == "" {
			continue
		}

		// Set field value based on type
		if err := l.setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s from env %s: %w", fieldType.Name, envTag, err)
		}
	}

	return nil
}

// buildEnvName builds environment variable name with prefix
func (l *Loader) buildEnvName(prefix, fieldName string) string {
	envName := strings.ToUpper(fieldName)

	if prefix != "" {
		envName = strings.ToUpper(prefix) + "_" + envName
	}

	if l.envPrefix != "" {
		envName = strings.ToUpper(l.envPrefix) + "_" + envName
	}

	return envName
}

// setFieldValue sets field value from string
func (l *Loader) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// MustLoad loads configuration and panics if it fails
func (l *Loader) MustLoad(cfg any) {
	if err := l.Load(cfg); err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvWithPrefix gets an environment variable with prefix
func GetEnvWithPrefix(prefix, key, defaultValue string) string {
	envKey := key
	if prefix != "" {
		envKey = strings.ToUpper(prefix + "_" + key)
	}
	return GetEnv(envKey, defaultValue)
}

// LoadConfig is a convenience function to load configuration
func LoadConfig(cfg any, opts ...Option) error {
	loader := NewLoader(opts...)
	return loader.Load(cfg)
}

// MustLoadConfig is a convenience function that panics on error
func MustLoadConfig(cfg any, opts ...Option) {
	loader := NewLoader(opts...)
	loader.MustLoad(cfg)
}
