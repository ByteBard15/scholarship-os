package profile

import (
	"net/http"

	"github.com/byte/scholarship-os/apps/api/pkg/response"
)

func (h *Handler) AgentListProfiles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.AgentListProfiles(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]ProfileResponse, 0, len(rows))
	for i := range rows {
		out = append(out, profileResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) AgentGetEffective(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	result, err := h.service.AgentEffectiveProfile(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}
func (h *Handler) AgentCreateDomain(w http.ResponseWriter, r *http.Request) {
	var req AgentCreateDomainRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.AgentCreateDomain(r.Context(), req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, profileResponse(result))
}
func (h *Handler) AgentUpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req AgentUpdateProfileRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.AgentUpdateProfile(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, profileResponse(result))
}
func (h *Handler) AgentCreateDescriptionOverride(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req CreateOverrideRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.AgentCreateDescriptionOverride(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, overrideResponse(result))
}
