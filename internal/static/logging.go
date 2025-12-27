package static

import (
	"net/http"

	"go.uber.org/zap"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Debug("request",
			zap.String("address", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))

		next.ServeHTTP(w, r)
	})
}
