package static

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func requestLogger(statusCode int, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("address", r.RemoteAddr).
			Msgf("--> %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Debug().
			Str("address", r.RemoteAddr).
			Msgf("<-- %d %s", statusCode, http.StatusText(statusCode))
	})
}
