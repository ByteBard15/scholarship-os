package middleware

import (
	"log/slog"
	"net/http"
	"time"

	appLogger "github.com/example/scholarship-os/apps/api/pkg/logger"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := chiMiddleware.GetReqID(r.Context())
			ctx := appLogger.WithRequestID(r.Context(), log, requestID)
			r = r.WithContext(ctx)

			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			started := time.Now()
			next.ServeHTTP(ww, r)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			appLogger.Info(r.Context(), "http request", "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", time.Since(started).Milliseconds())
		})
	}
}
