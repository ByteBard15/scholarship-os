package workflow

import (
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"net/http"
)

func (h *Handler) ListProposalRisk(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "proposalID")
	if !ok {
		return
	}
	items, err := h.service.ListProposalRisk(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, items, len(items))
}
func (h *Handler) CreateProposalRisk(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "proposalID")
	if !ok {
		return
	}
	var q CreateRiskAssessmentRequest
	if !h.decode(w, r, &q) {
		return
	}
	v, err := h.service.CreateProposalRisk(r.Context(), id, q)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, v)
}
