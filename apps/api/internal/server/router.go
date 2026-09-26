package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/example/scholarship-os/apps/api/internal/features/application"
	featureauth "github.com/example/scholarship-os/apps/api/internal/features/auth"
	"github.com/example/scholarship-os/apps/api/internal/features/catalog"
	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/example/scholarship-os/apps/api/internal/features/user"
	"github.com/example/scholarship-os/apps/api/internal/features/workflow"
	appmw "github.com/example/scholarship-os/apps/api/internal/middleware"
	"github.com/example/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func NewRouter(log *slog.Logger, sqlDB *sql.DB, db *gorm.DB, webOrigin string, authenticator *featureauth.Service, authHandler *featureauth.Handler, users *user.Handler, profiles *profile.Handler, catalogs *catalog.Handler, applications *application.Handler, workflows *workflow.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, appmw.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{webOrigin}, AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}, ExposedHeaders: []string{"X-Request-ID"}, AllowCredentials: false, MaxAge: 300}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := sqlDB.PingContext(r.Context()); err != nil {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}
		response.JSON(w, code, map[string]string{"status": status, "database": status})
	})
	r.Route("/api/v1", func(r chi.Router) {
		featureauth.MountPublicRoutes(r, authHandler)
		r.Group(func(r chi.Router) {
			r.Use(appmw.Authenticate(authenticator), appmw.RequireUserOrSystem, appmw.EnforceWorkspaceOwnership(db))
			featureauth.MountUserRoutes(r, authHandler)
			user.MountRoutes(r, users)
			profile.MountRoutes(r, profiles)
			catalog.MountRoutes(r, catalogs)
			application.MountRoutes(r, applications)
			workflow.MountRoutes(r, workflows)
		})
		r.Group(func(r chi.Router) {
			r.Use(appmw.Authenticate(authenticator), appmw.RequireSystem)
			featureauth.MountSystemRoutes(r, authHandler)
		})
		r.Group(func(r chi.Router) {
			r.Use(appmw.Authenticate(authenticator), appmw.EnforceWorkspaceOwnership(db))
			workflow.MountAgentRoutes(r, workflows, applications)
		})
	})
	return r
}
