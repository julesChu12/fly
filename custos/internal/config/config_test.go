package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigUsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "9090")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_DATABASE", "custos_test")
	t.Setenv("DB_CHARSET", "utf8mb4")
	t.Setenv("JWT_SECRET", "token-secret")
	t.Setenv("JWT_ACCESS_TTL", "30")
	t.Setenv("JWT_REFRESH_TTL", "2880")

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
	require.Equal(t, int64(30), int64(cfg.JWT.AccessTokenTTL.Minutes()))
	require.Equal(t, int64(2880), int64(cfg.JWT.RefreshTokenTTL.Minutes()))

	require.Equal(t, "tester:secret@tcp(db:3307)/custos_test?charset=utf8mb4&parseTime=True&loc=Local", cfg.Database.DSN())
}

func TestLoadConfigFailsWithoutSecret(t *testing.T) {
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_DATABASE", "custos_test")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	require.Error(t, err)
}
