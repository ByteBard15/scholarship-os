package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestWithRequestIDAddsAttribute(t *testing.T) {
	var output bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := WithRequestID(context.Background(), base, "request-123")

	Info(ctx, "handled request")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if got := record[RequestIDKey]; got != "request-123" {
		t.Fatalf("request ID = %v, want request-123", got)
	}
}

func TestFromContextUsesFallback(t *testing.T) {
	fallback := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	if got := FromContext(context.Background(), fallback); got != fallback {
		t.Fatal("expected fallback logger")
	}
}
