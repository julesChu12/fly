package mora

import (
	"github.com/julesChu12/fly/mora/pkg/config"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration using Mora config package
func LoadConfig(configPath string) (*viper.Viper, error) {
	loader := config.New()
	loader.WithYAML(configPath)
	return loader.Load()
}

// LoadConfigWithDefaults loads configuration with default values
func LoadConfigWithDefaults(configName string, configPaths []string) (*viper.Viper, error) {
	loader := config.New()

	// Build YAML paths from config name and paths
	yamlPaths := make([]string, 0, len(configPaths))
	for _, path := range configPaths {
		yamlPaths = append(yamlPaths, path+"/"+configName+".yaml")
	}

	loader.WithYAML(yamlPaths...)
	return loader.Load()
}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *config.Loader {
	return config.New()
}
