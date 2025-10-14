package mora

import (
	"github.com/julesChu12/fly/mora/pkg/config"
)

// LoadConfig loads configuration using Mora config package
func LoadConfig(configPath string, cfg interface{}) error {
	loader := config.NewViperLoader()
	return loader.Load(configPath, cfg)
}

// LoadConfigWithDefaults loads configuration with default values
func LoadConfigWithDefaults(configName string, configPaths []string, cfg interface{}) error {
	loader := config.NewViperLoader()

	// Try each path
	for _, path := range configPaths {
		if err := loader.Load(path+"/"+configName+".yaml", cfg); err == nil {
			return nil
		}
	}

	// Return default config if no file found
	return nil
}
