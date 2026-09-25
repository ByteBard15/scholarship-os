package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			started := time.Now()
			next.ServeHTTP(ww, r)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			log.InfoContext(r.Context(), "http request", "request_id", middleware.GetReqID(r.Context()), "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", time.Since(started).Milliseconds())
		})
	}
}
