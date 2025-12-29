package static

// MethodConfig for a specific endpoint
type MethodConfig struct {
	Method     string            `yaml:"method"`
	StatusCode int               `yaml:"status-code"`
	Body       string            `yaml:"body"`
	Headers    map[string]string `yaml:"headers"`
}

// SupportedMethods lists the supported methods for a given Endpoint
type SupportedMethods []string
