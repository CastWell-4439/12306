package logging

import (
	"context"
	"log/slog"
	"os"

	"ticketing/internal/common/tracing"
)

func New(service string, env string, version string) *slog.Logger {
	base := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	return base.With(
		"service", service,
		"env", env,
		"version", version,
	)
}

func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	traceID := tracing.TraceID(ctx)
	reqID := tracing.RequestID(ctx)
	if traceID == "" && reqID == "" {
		return logger
	}
	return logger.With("trace_id", traceID, "req_id", reqID)
}
