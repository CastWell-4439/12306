package middleware

import (
	"net/http"

	"ticketing-gozero/pkg/infra/tracing"
)

// WithRequestContext returns a go-zero compatible middleware that injects
// trace ID and request ID into the request context.
func WithRequestContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := tracing.EnsureTraceID(r.Header.Get("X-Trace-Id"))
		reqID := tracing.EnsureRequestID(r.Header.Get("X-Request-Id"))

		ctx := tracing.WithTraceAndRequestID(r.Context(), traceID, reqID)
		w.Header().Set("X-Trace-Id", traceID)
		w.Header().Set("X-Request-Id", reqID)

		next(w, r.WithContext(ctx))
	}
}

