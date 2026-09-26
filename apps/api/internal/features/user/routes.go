package user

import (
	"github.com/byte/scholarship-os/apps/api/internal/features/ownership"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router, h *Handler, owners *ownership.Middleware) {
	r.With(owners.User("get user")).Get("/users/{userID}", h.Get)
}
