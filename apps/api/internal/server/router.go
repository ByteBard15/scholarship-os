package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	featureauth "github.com/byte/scholarship-os/apps/api/internal/features/auth"
	"github.com/byte/scholarship-os/apps/api/internal/features/catalog"
	"github.com/byte/scholarship-os/apps/api/internal/features/ownership"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/byte/scholarship-os/apps/api/internal/features/workflow"
	"github.com/byte/scholarship-os/apps/api/internal/features/writing"
	appmw "github.com/byte/scholarship-os/apps/api/internal/middleware"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func NewRouter(log *slog.Logger, sqlDB *sql.DB, db *gorm.DB, webOrigin string, authenticator *featureauth.Service, authHandler *featureauth.Handler, users *user.Handler, profiles *profile.Handler, catalogs *catalog.Handler, applications *application.Handler, workflows *workflow.Handler, writings *writing.Handler) http.Handler {
	owners := ownership.NewMiddleware(ownership.NewGORMRepository(db))
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
			r.Use(appmw.Authenticate(authenticator), appmw.RequireUserOrSystem)
			featureauth.MountUserRoutes(r, authHandler)
			user.MountRoutes(r, users, owners)
			profile.MountRoutes(r, profiles, owners)
			catalog.MountRoutes(r, catalogs)
			application.MountRoutes(r, applications, owners)
			workflow.MountRoutes(r, workflows, owners)
			writing.MountRoutes(r, writings)
		})
		r.Group(func(r chi.Router) {
			r.Use(appmw.Authenticate(authenticator), appmw.RequireSystem)
			featureauth.MountSystemRoutes(r, authHandler)
		})
		r.Group(func(r chi.Router) {
			r.Use(appmw.Authenticate(authenticator))
			workflow.MountAgentRoutes(r, workflows, applications, owners)
			writing.MountAgentRoutes(r, writings)
		})
	})
	return r
}
