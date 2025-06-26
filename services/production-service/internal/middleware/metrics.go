package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shard-legends/production-service/pkg/metrics"
)

func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start).Seconds()
				status := strconv.Itoa(ww.Status())

				rctx := chi.RouteContext(r.Context())
				routePattern := r.URL.Path
				if rctx != nil && rctx.RoutePattern() != "" {
					routePattern = rctx.RoutePattern()
				}

				metrics.RecordHTTPRequest(r.Method, routePattern, status, duration)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
