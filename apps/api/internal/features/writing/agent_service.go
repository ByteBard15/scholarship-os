package writing

import (
	"context"
	"net/http"
	"strconv"
	"time"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/google/uuid"
)

type SampleMetadataQuery struct {
	Title        string
	DocumentType string
	Limit        int
	Offset       int
}
type SampleMetadataResponse struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID *uuid.UUID `json:"applicationId,omitempty"`
	Title         string     `json:"title"`
	DocumentType  string     `json:"documentType"`
	Status        string     `json:"status"`
	Origin        string     `json:"origin"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func requireWritingAgent(ctx context.Context) error {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return err
	}
	if actor.IsSystem() {
		return nil
	}
	if actor.Type != principal.PrincipalAgent {
		return principal.ErrAgentRequired
	}
	if !actor.HasScope(principal.ScopeWriting) {
		return principal.ErrAgentScopeRequired
	}
	if actor.UserID == nil {
		return principal.ErrForbidden
	}
	return nil
}
func (s *Service) AgentListSampleMetadata(ctx context.Context, q SampleMetadataQuery) ([]SampleMetadataResponse, error) {
	if err := requireWritingAgent(ctx); err != nil {
		return nil, err
	}
	actor, _ := principal.PrincipalFromContext(ctx)
	if actor.UserID == nil {
		return nil, ErrUserRequired
	}
	if q.Limit < 1 || q.Limit > 50 || q.Offset < 0 || q.Offset > 1000000 {
		return nil, ErrInvalidQuery
	}
	return s.repo.ListSampleMetadata(ctx, *actor.UserID, q)
}
func (h *Handler) AgentListSampleMetadata(w http.ResponseWriter, r *http.Request) {
	q := SampleMetadataQuery{Title: r.URL.Query().Get("title"), DocumentType: r.URL.Query().Get("documentType"), Limit: 20}
	for _, p := range []struct {
		name  string
		value *int
	}{{"limit", &q.Limit}, {"offset", &q.Offset}} {
		if raw := r.URL.Query().Get(p.name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				h.writeError(w, r, ErrInvalidQuery)
				return
			}
			*p.value = value
		}
	}
	out, err := h.service.AgentListSampleMetadata(r.Context(), q)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, out, len(out))
}
