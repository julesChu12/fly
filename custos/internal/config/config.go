package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	moracfg "github.com/julesChu12/fly/mora/pkg/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuth
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Charset  string
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (*Config, error) {
	v, err := moracfg.New().
		WithDotenv(".env").
		WithYAML("configs/custos.yaml").
		WithEnvPrefix("CUSTOS").
		Load()
	if err != nil {
		return nil, fmt.Errorf("load base config failed: %w", err)
	}

	setDefaults(v)

	if err := bindEnv(v); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(durationDecodeHook())); err != nil {
		return nil, fmt.Errorf("unmarshal to Config failed: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.port", "8080")
	v.SetDefault("app.env", "development")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "3306")
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "custos")
	v.SetDefault("database.charset", "utf8mb4")

	v.SetDefault("jwt.secretKey", "dev-secret-change-me")
	v.SetDefault("jwt.accessTokenTTL", "15m")
	v.SetDefault("jwt.refreshTokenTTL", "168h")

	// OAuth defaults
	v.SetDefault("oauth.stateKey", "dev-oauth-state-key-change-me")
	v.SetDefault("oauth.stateTTL", 600) // 10 minutes

	// Google OAuth defaults
	v.SetDefault("oauth.google.authURL", "https://accounts.google.com/o/oauth2/auth")
	v.SetDefault("oauth.google.tokenURL", "https://oauth2.googleapis.com/token")
	v.SetDefault("oauth.google.userInfoURL", "https://www.googleapis.com/oauth2/v2/userinfo")
	v.SetDefault("oauth.google.scopes", []string{"openid", "email", "profile"})

	// GitHub OAuth defaults
	v.SetDefault("oauth.github.authURL", "https://github.com/login/oauth/authorize")
	v.SetDefault("oauth.github.tokenURL", "https://github.com/login/oauth/access_token")
	v.SetDefault("oauth.github.userInfoURL", "https://api.github.com/user")
	v.SetDefault("oauth.github.scopes", []string{"user:email"})
}

func bindEnv(v *viper.Viper) error {
	bindings := map[string][]string{
		"app.port":                     {"CUSTOS_APP_PORT", "CUSTOS_PORT", "PORT"},
		"app.env":                      {"CUSTOS_APP_ENV", "APP_ENV"},
		"database.host":                {"CUSTOS_DB_HOST", "DB_HOST"},
		"database.port":                {"CUSTOS_DB_PORT", "DB_PORT"},
		"database.user":                {"CUSTOS_DB_USER", "DB_USER"},
		"database.password":            {"CUSTOS_DB_PASSWORD", "DB_PASSWORD"},
		"database.database":            {"CUSTOS_DB_DATABASE", "DB_DATABASE"},
		"database.charset":             {"CUSTOS_DB_CHARSET", "DB_CHARSET"},
		"jwt.secretKey":                {"CUSTOS_JWT_SECRET_KEY", "JWT_SECRET"},
		"jwt.accessTokenTTL":           {"CUSTOS_JWT_ACCESS_TOKEN_TTL", "JWT_ACCESS_TTL"},
		"jwt.refreshTokenTTL":          {"CUSTOS_JWT_REFRESH_TOKEN_TTL", "JWT_REFRESH_TTL"},
		"oauth.stateKey":               {"CUSTOS_OAUTH_STATE_KEY", "OAUTH_STATE_KEY"},
		"oauth.stateTTL":               {"CUSTOS_OAUTH_STATE_TTL", "OAUTH_STATE_TTL"},
		"oauth.google.clientID":        {"CUSTOS_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"},
		"oauth.google.clientSecret":    {"CUSTOS_GOOGLE_CLIENT_SECRET", "GOOGLE_CLIENT_SECRET"},
		"oauth.github.clientID":        {"CUSTOS_GITHUB_CLIENT_ID", "GITHUB_CLIENT_ID"},
		"oauth.github.clientSecret":    {"CUSTOS_GITHUB_CLIENT_SECRET", "GITHUB_CLIENT_SECRET"},
	}

	for key, envs := range bindings {
		args := append([]string{key}, envs...)
		if err := v.BindEnv(args...); err != nil {
			return fmt.Errorf("bind env for %s: %w", key, err)
		}
	}

	return nil
}

func durationDecodeHook() mapstructure.DecodeHookFunc {
	durationType := reflect.TypeOf(time.Duration(0))

	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != durationType {
			return data, nil
		}

		switch value := data.(type) {
		case string:
			parsed := strings.TrimSpace(value)
			if parsed == "" {
				return time.Duration(0), nil
			}
			if d, err := time.ParseDuration(parsed); err == nil {
				return d, nil
			}
			if minutes, err := strconv.Atoi(parsed); err == nil {
				return time.Duration(minutes) * time.Minute, nil
			}
			return nil, fmt.Errorf("invalid duration value %q", value)
		case int:
			return time.Duration(value) * time.Minute, nil
		case int8:
			return time.Duration(value) * time.Minute, nil
		case int16:
			return time.Duration(value) * time.Minute, nil
		case int32:
			return time.Duration(value) * time.Minute, nil
		case int64:
			return time.Duration(value) * time.Minute, nil
		case uint:
			return time.Duration(value) * time.Minute, nil
		case uint8:
			return time.Duration(value) * time.Minute, nil
		case uint16:
			return time.Duration(value) * time.Minute, nil
		case uint32:
			return time.Duration(value) * time.Minute, nil
		case uint64:
			return time.Duration(value) * time.Minute, nil
		case float32:
			return time.Duration(int64(value)) * time.Minute, nil
		case float64:
			return time.Duration(int64(value)) * time.Minute, nil
		default:
			return data, nil
		}
	}
}

func validate(cfg *Config) error {
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("jwt.secretKey is required")
	}
	if cfg.App.Env != "development" && cfg.JWT.SecretKey == "dev-secret-change-me" {
		return fmt.Errorf("in %s env, jwt.secretKey must not be the default value", cfg.App.Env)
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if cfg.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("jwt.accessTokenTTL must be greater than zero")
	}
	if cfg.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("jwt.refreshTokenTTL must be greater than zero")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

func (c *Config) IsDev() bool {
	return c.App.Env == "development"
}

func (c *Config) Env() string {
	return c.App.Env
}
