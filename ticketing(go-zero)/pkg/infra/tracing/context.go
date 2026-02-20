package tracing

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	reqIDKey   contextKey = "req_id"
)

func WithTraceAndRequestID(ctx context.Context, traceID string, reqID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, reqIDKey, reqID)
	return ctx
}

func TraceID(ctx context.Context) string {
	v, _ := ctx.Value(traceIDKey).(string)
	return v
}

func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(reqIDKey).(string)
	return v
}

func EnsureTraceID(v string) string {
	if v != "" {
		return v
	}
	return uuid.NewString()
}

func EnsureRequestID(v string) string {
	if v != "" {
		return v
	}
	return uuid.NewString()
}

