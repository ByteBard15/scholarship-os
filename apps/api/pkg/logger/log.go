package logger

import (
	"context"
	"log/slog"
)

const RequestIDKey = "request_id"

type contextKey struct{}

// WithContext stores a logger for the lifetime of ctx. A nil logger falls back
// to slog.Default, so request processing never fails because logging was not
// explicitly configured.
func WithContext(ctx context.Context, requestLogger *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if requestLogger == nil {
		requestLogger = slog.Default()
	}
	return context.WithValue(ctx, contextKey{}, requestLogger)
}

// WithRequestID derives a request-scoped logger and stores it in ctx. The ID is
// attached once here and is inherited by every log written through this package.
func WithRequestID(ctx context.Context, base *slog.Logger, requestID string) context.Context {
	if base == nil {
		base = slog.Default()
	}
	if requestID != "" {
		base = base.With(slog.String(RequestIDKey, requestID))
	}
	return WithContext(ctx, base)
}

// FromContext returns the request-scoped logger when available. The first
// non-nil fallback is used outside an HTTP request; otherwise slog.Default is
// returned.
func FromContext(ctx context.Context, fallback ...*slog.Logger) *slog.Logger {
	if ctx != nil {
		if value, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && value != nil {
			return value
		}
	}
	for _, candidate := range fallback {
		if candidate != nil {
			return candidate
		}
	}
	return slog.Default()
}

// Logger is retained as the concise accessor for call sites that do not need a
// custom fallback.
func Logger(ctx context.Context) *slog.Logger {
	return FromContext(ctx)
}

func Debug(ctx context.Context, message string, args ...any) {
	ctx = nonNilContext(ctx)
	Logger(ctx).DebugContext(ctx, message, args...)
}

func Info(ctx context.Context, message string, args ...any) {
	ctx = nonNilContext(ctx)
	Logger(ctx).InfoContext(ctx, message, args...)
}

func Warn(ctx context.Context, message string, args ...any) {
	ctx = nonNilContext(ctx)
	Logger(ctx).WarnContext(ctx, message, args...)
}

func Error(ctx context.Context, message string, args ...any) {
	ctx = nonNilContext(ctx)
	Logger(ctx).ErrorContext(ctx, message, args...)
}

func Log(ctx context.Context, level slog.Level, message string, args ...any) {
	ctx = nonNilContext(ctx)
	Logger(ctx).Log(ctx, level, message, args...)
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
