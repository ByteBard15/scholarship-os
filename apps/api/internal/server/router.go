package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	appmw "github.com/example/scholarship-os/apps/api/internal/middleware"
	"github.com/example/scholarship-os/apps/api/internal/profile"
	"github.com/example/scholarship-os/apps/api/internal/response"
	"github.com/example/scholarship-os/apps/api/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(log *slog.Logger, db *sql.DB, webOrigin string, users *user.Handler, profiles *profile.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, appmw.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{webOrigin}, AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}, ExposedHeaders: []string{"X-Request-ID"}, AllowCredentials: false, MaxAge: 300}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := db.PingContext(r.Context()); err != nil {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}
		response.JSON(w, code, map[string]string{"status": status, "database": status})
	})
	r.Route("/api/v1", func(r chi.Router) {
		user.MountRoutes(r, users)
		profile.MountRoutes(r, profiles)
	})
	return r
}
