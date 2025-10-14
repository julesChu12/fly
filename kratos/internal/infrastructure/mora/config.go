package mora

import (
	"github.com/julesChu12/fly/mora/pkg/config"
)

// LoadConfig loads configuration using Mora config package
func LoadConfig(configPath string, cfg interface{}) error {
	loader := config.New()
	v, err := loader.WithYAML(configPath).Load()
	if err != nil {
		return err
	}
	return v.Unmarshal(cfg)
}

// LoadConfigWithDefaults loads configuration with default values
func LoadConfigWithDefaults(configName string, configPaths []string, cfg interface{}) error {
	loader := config.New()

	// Try each path
	for _, path := range configPaths {
		yamlPath := path + "/" + configName + ".yaml"
		v, err := loader.WithYAML(yamlPath).Load()
		if err == nil {
			return v.Unmarshal(cfg)
		}
	}

	// Return default config if no file found
	return nil
}
