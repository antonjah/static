package static

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

// Endpoints represents one or more Endpoint that is appended to the listener
type Endpoints struct {
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint represents one static Endpoint that is appended to the listener
type Endpoint struct {
	Path       string            `yaml:"path"`
	StatusCode int               `yaml:"status-code"`
	Body       string            `yaml:"body"`
	Headers    map[string]string `yaml:"headers"`
}

// Validate makes sure that the required fields are set for a specific Endpoint
func (e Endpoint) Validate() error {
	if e.Path == "" {
		return errors.New("missing path")
	}

	if e.StatusCode < 100 || e.StatusCode > 599 {
		return fmt.Errorf("invalid status-code: %d", e.StatusCode)
	}

	return nil
}

// ServeHTTP responds to an HTTP request using the pre-configured
// content-type, body, and headers
func (e Endpoint) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// append header(s)
	for key, val := range e.Headers {
		w.Header().Add(key, val)
	}

	// write status code
	w.WriteHeader(e.StatusCode)

	// write body
	if _, err := w.Write([]byte(e.Body)); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}
