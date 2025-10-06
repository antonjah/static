package static

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("address", r.RemoteAddr).
			Msgf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
