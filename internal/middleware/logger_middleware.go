package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sunriseex/test_wallet/internal/logger"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(w, r)
		logger.Log.Infof(
			"[%s] %s %d %s",
			r.Method,
			r.URL.Path,
			rw.Status(),
			time.Since(start))
	})
}
