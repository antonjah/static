package static

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"JSONTypeAndBody", "application/json", `{"key": "value"}`},
		{"PLAINTypeAndBody", "text/plain", "test"},
		{"NoTypeAndBody", "", `{"key": "value"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType = tt.contentType
			responseBody = tt.body

			w := httptest.NewRecorder()
			handler().ServeHTTP(w, &http.Request{})

			if gotContentType := w.Header().Get("content-type"); gotContentType != tt.contentType {
				t.Errorf("content-type = %v, want %v", gotContentType, tt.contentType)
			}

			if gotBody := w.Body.String(); gotBody != tt.body {
				t.Errorf("body = %v, want %v", gotBody, tt.body)
			}
		})
	}
}
