package profile

import (
	"net/http"

	"github.com/example/scholarship-os/apps/api/pkg/response"
)

func (h *Handler) GetLineage(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.Lineage(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]LineageItemResponse, 0, len(rows))
	for _, item := range rows {
		out = append(out, LineageItemResponse{ID: item.ID, Name: item.Name, ProfileType: item.ProfileType})
	}
	response.Data(w, http.StatusOK, out)
}
func (h *Handler) GetEffective(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	result, err := h.service.ResolveEffectiveProfile(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}
func (h *Handler) GetCompleteness(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	result, err := h.service.Completeness(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}
func (h *Handler) Compare(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	otherID, ok := pathID(w, r, "otherProfileID")
	if !ok {
		return
	}
	result, err := h.service.Compare(r.Context(), profileID, otherID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}

func (h *Handler) CreateOverride(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req CreateOverrideRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.CreateOverride(r.Context(), profileID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, overrideResponse(result))
}
func (h *Handler) ListOverrides(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.ListOverrides(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]OverrideResponse, 0, len(rows))
	for i := range rows {
		out = append(out, overrideResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) UpdateOverride(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	overrideID, ok := pathID(w, r, "overrideID")
	if !ok {
		return
	}
	var req UpdateOverrideRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.UpdateOverride(r.Context(), profileID, overrideID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, overrideResponse(result))
}
func (h *Handler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	overrideID, ok := pathID(w, r, "overrideID")
	if !ok {
		return
	}
	if err := h.service.DeleteOverride(r.Context(), profileID, overrideID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req CreateSnapshotRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.CreateSnapshot(r.Context(), profileID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, snapshotResponse(result, true))
}
func (h *Handler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.ListSnapshots(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]SnapshotResponse, 0, len(rows))
	for i := range rows {
		out = append(out, snapshotResponse(&rows[i], false))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	snapshotID, ok := pathID(w, r, "snapshotID")
	if !ok {
		return
	}
	result, err := h.service.GetSnapshot(r.Context(), profileID, snapshotID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, snapshotResponse(result, true))
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.service.maxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(h.service.maxUploadBytes + 1024*1024); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_DOCUMENT", "invalid or oversized multipart upload")
		return
	}
	documentType := r.FormValue("documentType")
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_DOCUMENT", "file is required")
		return
	}
	defer file.Close()
	result, err := h.service.UploadDocument(r.Context(), profileID, documentType, header.Filename, header.Header.Get("Content-Type"), file)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, documentResponse(result))
}
func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.ListDocuments(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]DocumentResponse, 0, len(rows))
	for i := range rows {
		out = append(out, documentResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	documentID, ok := pathID(w, r, "documentID")
	if !ok {
		return
	}
	result, err := h.service.GetDocument(r.Context(), profileID, documentID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, documentResponse(result))
}
func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	documentID, ok := pathID(w, r, "documentID")
	if !ok {
		return
	}
	if err := h.service.DeleteDocument(r.Context(), profileID, documentID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateImport(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req CreateImportRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.CreateImport(r.Context(), profileID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, importResponse(result))
}
func (h *Handler) ListImports(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.ListImports(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]ImportResponse, 0, len(rows))
	for i := range rows {
		out = append(out, importResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) GetImport(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	importID, ok := pathID(w, r, "importID")
	if !ok {
		return
	}
	result, err := h.service.GetImport(r.Context(), profileID, importID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, importResponse(result))
}
func (h *Handler) ListImportCandidates(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	importID, ok := pathID(w, r, "importID")
	if !ok {
		return
	}
	rows, err := h.service.ListImportCandidates(r.Context(), profileID, importID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, rows, len(rows))
}
func (h *Handler) ReviewCandidate(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	importID, ok := pathID(w, r, "importID")
	if !ok {
		return
	}
	candidateID, ok := pathID(w, r, "candidateID")
	if !ok {
		return
	}
	var req CandidateActionRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.ReviewCandidate(r.Context(), profileID, importID, candidateID, req.Action)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}
func (h *Handler) ApplyImport(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	importID, ok := pathID(w, r, "importID")
	if !ok {
		return
	}
	result, err := h.service.ApplyImport(r.Context(), profileID, importID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, importResponse(result))
}

func (h *Handler) CreateEvidence(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req CreateEvidenceRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	result, err := h.service.CreateEvidence(r.Context(), profileID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, evidenceResponse(result))
}
func (h *Handler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	profileID, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	rows, err := h.service.ListEvidence(r.Context(), profileID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]EvidenceResponse, 0, len(rows))
	for i := range rows {
		out = append(out, evidenceResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}

func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if !h.decode(w, r, dst) {
		return false
	}
	if err := h.validate.Struct(dst); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return false
	}
	return true
}
