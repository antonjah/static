package static

import (
	"net/http"
)

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", contentType)
		_, _ = w.Write([]byte(responseBody))
	}
}
