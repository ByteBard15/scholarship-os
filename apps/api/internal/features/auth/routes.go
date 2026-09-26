package auth

import "github.com/go-chi/chi/v5"

func MountPublicRoutes(r chi.Router, handler *Handler) {
	r.Post("/auth/login", handler.Login)
}

func MountUserRoutes(r chi.Router, handler *Handler) {
	r.Post("/auth/logout", handler.Logout)
	r.Get("/auth/me", handler.Me)
	r.Post("/auth/change-password", handler.ChangePassword)
}

func MountSystemRoutes(r chi.Router, handler *Handler) {
	r.Post("/auth/register", handler.Register)
	r.Get("/admin/users", handler.ListUsers)
	r.Get("/admin/users/{userID}", handler.GetUser)
	r.Patch("/admin/users/{userID}", handler.UpdateUser)
	r.Post("/admin/users/{userID}/disable", handler.SetUserActive(false))
	r.Post("/admin/users/{userID}/enable", handler.SetUserActive(true))
	r.Post("/admin/users/{userID}/reset-password", handler.ResetPassword)
	r.Get("/admin/agent-credentials", handler.ListAgentCredentials)
	r.Post("/admin/agent-credentials", handler.CreateAgentCredential)
	r.Post("/admin/agent-credentials/{credentialID}/revoke", handler.RevokeAgentCredential)
}
