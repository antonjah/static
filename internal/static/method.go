package static

import (
	"github.com/rs/zerolog"
)

// Method for a specific endpoint
type Method struct {
	Method     string            `yaml:"method"`
	StatusCode int               `yaml:"status-code"`
	Body       string            `yaml:"body"`
	Headers    map[string]string `yaml:"headers"`
}

// SupportedMethods lists the supported methods for a given Endpoint
type SupportedMethods []string

// MarshalZerologArray implements zerolog.LogArrayMarshaler
func (s SupportedMethods) MarshalZerologArray(a *zerolog.Array) {
	for _, method := range s {
		a.Str(method)
	}
}
