package writing

import (
	appmw "github.com/byte/scholarship-os/apps/api/internal/middleware"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router, handler *Handler) {
	r.Get("/writing-samples", handler.ListSamples)
	r.Post("/writing-samples", handler.CreateSample)
	r.Get("/writing-samples/{sampleID}", handler.GetSample)
	r.Patch("/writing-samples/{sampleID}", handler.UpdateSample)
	r.Delete("/writing-samples/{sampleID}", handler.DeleteSample)
	r.Get("/writing-tags", handler.ListTags)
	r.Post("/writing-tags", handler.CreateTag)
	r.Delete("/writing-tags/{tagID}", handler.DeleteTag)
}

func MountAgentRoutes(r chi.Router, handler *Handler) {
	writingScope := appmw.RequireAgentScope(principal.ScopeWriting)
	r.With(writingScope).Get("/agent/writing-tags", handler.AgentListTags)
	r.With(writingScope).Post("/agent/writing-samples/match", handler.AgentMatchSamples)
	r.With(writingScope).Post("/agent/writing-samples", handler.AgentCreateGeneratedSample)
}
