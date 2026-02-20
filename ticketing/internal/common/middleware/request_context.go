package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketing/internal/common/tracing"
)

func WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := tracing.EnsureTraceID(r.Header.Get("X-Trace-Id"))
		reqID := tracing.EnsureRequestID(r.Header.Get("X-Request-Id"))

		ctx := tracing.WithTraceAndRequestID(r.Context(), traceID, reqID)
		w.Header().Set("X-Trace-Id", traceID)
		w.Header().Set("X-Request-Id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithRequestContextGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := tracing.EnsureTraceID(c.GetHeader("X-Trace-Id"))
		reqID := tracing.EnsureRequestID(c.GetHeader("X-Request-Id"))

		ctx := tracing.WithTraceAndRequestID(c.Request.Context(), traceID, reqID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-Id", traceID)
		c.Header("X-Request-Id", reqID)

		c.Next()
	}
}
