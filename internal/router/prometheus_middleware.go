package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		status := strconv.Itoa(ww.Status())

		httpRequestsTotal.WithLabelValues(
			r.Method,
			routePattern,
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			r.Method,
			routePattern,
			status,
		).Observe(time.Since(start).Seconds())
	})
}
