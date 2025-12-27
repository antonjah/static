package static

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type StaticAPIs struct {
	StaticAPIs []StaticAPI `yaml:"staticapis"`
}

type StaticAPI struct {
	Path    string         `yaml:"path"`
	Methods []MethodConfig `yaml:"methods"`

	SupportedMethods SupportedMethods
}

func (e *StaticAPI) MethodFromRequest(req *http.Request) MethodConfig {
	for _, method := range e.Methods {
		if strings.Compare(strings.ToLower(method.Method), strings.ToLower(req.Method)) == 0 {
			return method
		}
	}
	return MethodConfig{}
}

func (e *StaticAPI) SetSupported() {
	for _, method := range e.Methods {
		e.SupportedMethods = append(e.SupportedMethods, method.Method)
	}
}

func (e *StaticAPI) Validate() error {
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

func (e *StaticAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
		zap.L().Error("failed to write response", zap.Error(err))
	}
}
