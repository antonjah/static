package static

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

// Endpoints represents one or more Endpoint that is appended to the listener
type Endpoints struct {
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint represents one static Endpoint that is appended to the listener
type Endpoint struct {
	Path    string   `yaml:"path"`
	Methods []Method `yaml:"methods"`

	SupportedMethods SupportedMethods
}

// MethodFromRequest is a helper function to get the method dynamically from a request
func (e *Endpoint) MethodFromRequest(req *http.Request) Method {
	for _, method := range e.Methods {
		if strings.Compare(strings.ToLower(method.Method), strings.ToLower(req.Method)) == 0 {
			return method
		}
	}
	return Method{}
}

// SetSupported creates a list of the supported methods for the given endpoint
// This is then used to validate requests in ServeHTTP
func (e *Endpoint) SetSupported() {
	for _, method := range e.Methods {
		e.SupportedMethods = append(e.SupportedMethods, method.Method)
	}
}

// Validate makes sure that the required fields are set for a specific Endpoint
func (e *Endpoint) Validate() error {
	// validate path
	if e.Path == "" {
		return errors.New("missing path")
	}

	// validate status code for methods
	for _, method := range e.Methods {
		if method.StatusCode < 100 || method.StatusCode > 599 {
			return fmt.Errorf("invalid status-code for method %s: %d", e.Path, method.StatusCode)
		}
	}

	return nil
}

// ServeHTTP responds to an HTTP request using the pre-configured
// content-type, body, and headers
func (e *Endpoint) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// make sure the HTTP method is supported
	if !slices.Contains(e.SupportedMethods, req.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get the requested method
	method := e.MethodFromRequest(req)

	// append header(s)
	for key, val := range method.Headers {
		w.Header().Add(key, val)
	}

	// write status code
	w.WriteHeader(method.StatusCode)

	// write body
	if _, err := w.Write([]byte(method.Body)); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}
