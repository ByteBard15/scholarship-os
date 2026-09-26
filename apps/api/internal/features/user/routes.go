package user

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, h *Handler) {
	r.Get("/users/{userID}", h.Get)
}
