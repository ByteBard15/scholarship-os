package profile

import (
	"net/http"

	"github.com/byte/scholarship-os/apps/api/pkg/response"
)

const maxProfileBundleBytes = 5 << 20

func (h *Handler) ExportBundle(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	bundle, err := h.service.ExportBundle(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, bundle)
}

func (h *Handler) ImportBundle(w http.ResponseWriter, r *http.Request) {
	userID, ok := pathID(w, r, "userID")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileBundleBytes)
	var request ImportProfileBundleRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	profile, err := h.service.ImportBundle(r.Context(), userID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, fullResponse(profile))
}
