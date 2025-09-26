package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Loader struct {
	dotenvPaths []string
	yamlPaths   []string
	envPrefix   string
}

func New() *Loader {
	return &Loader{}
}

// NewLoader is kept for backward compatibility with older code.
func NewLoader(_ ...any) *Loader {
	return New()
}

func (l *Loader) WithDotenv(paths ...string) *Loader {
	if len(paths) == 0 {
		if len(l.dotenvPaths) == 0 {
			l.dotenvPaths = []string{".env"}
		}
		return l
	}

	l.dotenvPaths = append(l.dotenvPaths, paths...)
	return l
}

func (l *Loader) WithYAML(paths ...string) *Loader {
	if len(paths) == 0 {
		return l
	}

	l.yamlPaths = append(l.yamlPaths, paths...)
	return l
}

func (l *Loader) WithEnvPrefix(prefix string) *Loader {
	l.envPrefix = prefix
	return l
}

func (l *Loader) Load() (*viper.Viper, error) {
	v := viper.New()

	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	if l.envPrefix != "" {
		v.SetEnvPrefix(l.envPrefix)
	}
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)

	if err := l.applyDotenv(); err != nil {
		return nil, err
	}

	if err := l.mergeYAML(v); err != nil {
		return nil, err
	}

	return v, nil
}

func (l *Loader) MustLoad() *viper.Viper {
	v, err := l.Load()
	if err != nil {
		panic(err)
	}
	return v
}

func (l *Loader) applyDotenv() error {
	if len(l.dotenvPaths) == 0 {
		return nil
	}

	for _, path := range l.dotenvPaths {
		if path == "" {
			continue
		}

		envMap, err := godotenv.Read(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("read dotenv %s: %w", path, err)
		}

		for key, value := range envMap {
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s from %s: %w", key, path, err)
			}
		}
	}

	return nil
}

func (l *Loader) mergeYAML(v *viper.Viper) error {
	if len(l.yamlPaths) == 0 {
		return nil
	}

	for _, path := range l.yamlPaths {
		if path == "" {
			continue
		}

		v.SetConfigFile(path)
		if err := v.MergeInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("merge config file %s: %w", path, err)
		}
	}

	return nil
}
