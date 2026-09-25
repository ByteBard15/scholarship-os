package user

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, h *Handler) {
	r.Post("/users", h.Create)
	r.Get("/users/{userID}", h.Get)
}
