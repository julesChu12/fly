package observability

// Config holds the observability configuration
type Config struct {
	ServiceName  string  `json:"service_name" yaml:"service_name"`   // Service name for traces
	ExporterURL  string  `json:"exporter_url" yaml:"exporter_url"`   // OTLP endpoint URL
	SampleRatio  float64 `json:"sample_ratio" yaml:"sample_ratio"`   // Sampling ratio (0.0 to 1.0)
	Environment  string  `json:"environment" yaml:"environment"`     // Environment (dev, staging, prod)
	ExporterType string  `json:"exporter_type" yaml:"exporter_type"` // Exporter type: otlp, jaeger, stdout
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:  "mora-service",
		ExporterURL:  "http://localhost:4317",
		SampleRatio:  1.0,
		Environment:  "development",
		ExporterType: "otlp",
	}
}
