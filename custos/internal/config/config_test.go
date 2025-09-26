package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigUsesPrefixedEnvOverrides(t *testing.T) {
	t.Setenv("CUSTOS_APP_ENV", "test")
	t.Setenv("CUSTOS_APP_PORT", "9090")
	t.Setenv("CUSTOS_DB_HOST", "db")
	t.Setenv("CUSTOS_DB_PORT", "3307")
	t.Setenv("CUSTOS_DB_USER", "tester")
	t.Setenv("CUSTOS_DB_PASSWORD", "secret")
	t.Setenv("CUSTOS_DB_DATABASE", "custos_test")
	t.Setenv("CUSTOS_DB_CHARSET", "utf8mb4")
	t.Setenv("CUSTOS_JWT_SECRET_KEY", "token-secret")
	t.Setenv("CUSTOS_JWT_ACCESS_TOKEN_TTL", "30m")
	t.Setenv("CUSTOS_JWT_REFRESH_TOKEN_TTL", "336h")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "test", cfg.App.Env)
	require.Equal(t, "9090", cfg.App.Port)
	require.Equal(t, "db", cfg.Database.Host)
	require.Equal(t, "3307", cfg.Database.Port)
	require.Equal(t, "tester", cfg.Database.User)
	require.Equal(t, "secret", cfg.Database.Password)
	require.Equal(t, "custos_test", cfg.Database.Database)
	require.Equal(t, "utf8mb4", cfg.Database.Charset)
	require.Equal(t, "token-secret", cfg.JWT.SecretKey)
	require.Equal(t, 30*time.Minute, cfg.JWT.AccessTokenTTL)
	require.Equal(t, 336*time.Hour, cfg.JWT.RefreshTokenTTL)

	require.Equal(t, "tester:secret@tcp(db:3307)/custos_test?charset=utf8mb4&parseTime=True&loc=Local", cfg.Database.DSN())
}

func TestLoadConfigSupportsLegacyEnvFallbacks(t *testing.T) {
	t.Setenv("APP_ENV", "legacy")
	t.Setenv("PORT", "8088")
	t.Setenv("DB_HOST", "legacy-db")
	t.Setenv("DB_PORT", "3308")
	t.Setenv("DB_USER", "legacy-user")
	t.Setenv("DB_DATABASE", "legacy_custos")
	t.Setenv("JWT_SECRET", "legacy-secret")
	t.Setenv("JWT_ACCESS_TTL", "45")
	t.Setenv("JWT_REFRESH_TTL", "1440")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "legacy", cfg.App.Env)
	require.Equal(t, "8088", cfg.App.Port)
	require.Equal(t, "legacy-db", cfg.Database.Host)
	require.Equal(t, 45*time.Minute, cfg.JWT.AccessTokenTTL)
	require.Equal(t, 1440*time.Minute, cfg.JWT.RefreshTokenTTL)
}

func TestLoadConfigFailsWithoutSecret(t *testing.T) {
	t.Setenv("CUSTOS_DB_USER", "tester")
	t.Setenv("CUSTOS_DB_DATABASE", "custos_test")
	t.Setenv("CUSTOS_JWT_SECRET_KEY", "")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadConfigRejectsDefaultSecretInProd(t *testing.T) {
	t.Setenv("CUSTOS_APP_ENV", "production")
	t.Setenv("CUSTOS_DB_USER", "tester")
	t.Setenv("CUSTOS_DB_DATABASE", "custos")

	_, err := Load()
	require.Error(t, err)
}
