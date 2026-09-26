package profile

import "encoding/json"

func overrideResponse(v *ProfileOverride) OverrideResponse {
	return OverrideResponse{ID: v.ID, ProfileID: v.ProfileID, EntityType: v.EntityType, EntityID: v.EntityID, FieldName: v.FieldName, OverrideType: v.OverrideType, Value: json.RawMessage(v.Value), Reason: v.Reason, Source: v.Source, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func snapshotResponse(v *ProfileSnapshot, includeSnapshot bool) SnapshotResponse {
	out := SnapshotResponse{ID: v.ID, ProfileID: v.ProfileID, Version: v.Version, CreatedAt: v.CreatedAt, CreatedBy: v.CreatedBy, Reason: v.Reason}
	if includeSnapshot {
		out.Snapshot = json.RawMessage(v.Snapshot)
	}
	return out
}
func documentResponse(v *ProfileDocument) DocumentResponse {
	return DocumentResponse{ID: v.ID, ProfileID: v.ProfileID, DocumentType: v.DocumentType, OriginalFilename: v.OriginalFilename, StorageProvider: v.StorageProvider, MimeType: v.MimeType, FileSize: v.FileSize, SHA256: v.SHA256, Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func importResponse(v *ProfileImport) ImportResponse {
	return ImportResponse{ID: v.ID, ProfileID: v.ProfileID, DocumentID: v.DocumentID, ImportType: v.ImportType, Status: v.Status, ExtractedData: json.RawMessage(v.ExtractedData), ExtractionProvider: v.ExtractionProvider, ExtractionModel: v.ExtractionModel, ErrorMessage: v.ErrorMessage, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, CompletedAt: v.CompletedAt}
}
func candidateResponse(v *ProfileImportCandidate) ImportCandidateResponse {
	return ImportCandidateResponse{ID: v.ID, ImportID: v.ImportID, SectionType: v.SectionType, CandidateData: json.RawMessage(v.CandidateData), SourceText: v.SourceText, Confidence: v.Confidence, Status: v.Status, MatchedEntityID: v.MatchedEntityID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func evidenceResponse(v *ProfileEvidence) EvidenceResponse {
	return EvidenceResponse{ID: v.ID, ProfileID: v.ProfileID, EntityType: v.EntityType, EntityID: v.EntityID, FieldName: v.FieldName, EvidenceType: v.EvidenceType, DocumentID: v.DocumentID, ImportID: v.ImportID, SourceURL: v.SourceURL, SourceText: v.SourceText, Confidence: v.Confidence, CreatedAt: v.CreatedAt}
}
