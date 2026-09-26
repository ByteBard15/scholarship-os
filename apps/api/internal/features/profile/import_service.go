package profile

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)
var documentTypes = map[string]bool{"cv": true, "transcript": true, "certificate": true, "publication": true, "portfolio": true, "other": true}
var documentMIMEs = map[string]string{"application/pdf": ".pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx", "text/plain": ".txt"}

func (s *Service) UploadDocument(ctx context.Context, profileID uuid.UUID, documentType, filename, mimeType string, reader io.Reader) (*ProfileDocument, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	if !documentTypes[documentType] {
		return nil, ErrInvalidDocumentType
	}
	expectedExtension, ok := documentMIMEs[mimeType]
	if !ok {
		return nil, ErrInvalidDocument
	}
	safeName := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	if safeName == "." || safeName == "" || !strings.EqualFold(filepath.Ext(safeName), expectedExtension) {
		return nil, ErrInvalidDocument
	}
	buffered := bufio.NewReader(reader)
	header, err := buffered.Peek(4)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, ErrInvalidDocument
	}
	if !validMagic(mimeType, header) {
		return nil, ErrInvalidDocument
	}
	saved, err := s.files.Save(ctx, safeName, mimeType, io.LimitReader(buffered, s.maxUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if saved.Size == 0 || saved.Size > s.maxUploadBytes {
		_ = s.files.Delete(ctx, saved.Key)
		return nil, ErrInvalidDocument
	}
	hash := saved.SHA256
	document := &ProfileDocument{ProfileID: profileID, DocumentType: documentType, OriginalFilename: safeName, StorageProvider: "local", StorageKey: saved.Key, MimeType: mimeType, FileSize: saved.Size, SHA256: &hash, Status: "ready"}
	if err = s.workflow.CreateDocument(ctx, document); err != nil {
		_ = s.files.Delete(ctx, saved.Key)
		return nil, err
	}
	return document, nil
}
func validMagic(mimeType string, header []byte) bool {
	switch mimeType {
	case "application/pdf":
		return bytes.HasPrefix(header, []byte("%PDF"))
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return bytes.HasPrefix(header, []byte("PK"))
	case "text/plain":
		return !bytes.Contains(header, []byte{0})
	}
	return false
}
func (s *Service) ListDocuments(ctx context.Context, profileID uuid.UUID) ([]ProfileDocument, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.ListDocuments(ctx, profileID)
}
func (s *Service) GetDocument(ctx context.Context, profileID, id uuid.UUID) (*ProfileDocument, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.GetDocument(ctx, profileID, id)
}
func (s *Service) DeleteDocument(ctx context.Context, profileID, id uuid.UUID) error {
	document, err := s.GetDocument(ctx, profileID, id)
	if err != nil {
		return err
	}
	if err = s.workflow.DeleteDocument(ctx, profileID, id); err != nil {
		return err
	}
	return s.files.Delete(ctx, document.StorageKey)
}

func (s *Service) CreateImport(ctx context.Context, profileID uuid.UUID, req CreateImportRequest) (*ProfileImport, error) {
	if s.files == nil || s.textExtractor == nil || s.extractor == nil {
		return nil, errors.New("import workflow is not configured")
	}
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	document, err := s.workflow.GetDocument(ctx, profileID, req.DocumentID)
	if err != nil {
		return nil, err
	}
	if document.DocumentType != "cv" {
		return nil, ErrInvalidDocumentType
	}
	imp := &ProfileImport{ProfileID: profileID, DocumentID: document.ID, ImportType: "cv", Status: "extracting"}
	if err = s.workflow.CreateImport(ctx, imp); err != nil {
		return nil, err
	}
	file, err := s.files.Open(ctx, document.StorageKey)
	if err != nil {
		return s.failImport(ctx, imp, err)
	}
	defer file.Close()
	rawText, err := s.textExtractor.Extract(ctx, document, file)
	if err != nil {
		return s.failImport(ctx, imp, err)
	}
	extracted, err := s.extractor.Extract(ctx, rawText)
	if err != nil {
		return s.failImport(ctx, imp, err)
	}
	serialized, err := json.Marshal(extracted)
	if err != nil {
		return s.failImport(ctx, imp, err)
	}
	provider, model := s.extractor.Provider(), s.extractor.Model()
	imp.RawText = &rawText
	imp.ExtractedData = JSON(serialized)
	imp.ExtractionProvider = &provider
	imp.ExtractionModel = &model
	imp.Status = "review_required"
	candidates, err := s.buildCandidates(ctx, profileID, imp.ID, extracted)
	if err != nil {
		return s.failImport(ctx, imp, err)
	}
	if err = s.workflow.WithinTransaction(ctx, func(_ Repository, _ SectionRepository, workflow WorkflowRepository) error {
		if err := workflow.UpdateImport(ctx, imp); err != nil {
			return err
		}
		return workflow.CreateCandidates(ctx, candidates)
	}); err != nil {
		return nil, err
	}
	return imp, nil
}
func (s *Service) failImport(ctx context.Context, imp *ProfileImport, cause error) (*ProfileImport, error) {
	message := cause.Error()
	now := time.Now().UTC()
	imp.Status = "failed"
	imp.ErrorMessage = &message
	imp.CompletedAt = &now
	if err := s.workflow.UpdateImport(ctx, imp); err != nil {
		return nil, err
	}
	return imp, nil
}
func (s *Service) ListImports(ctx context.Context, profileID uuid.UUID) ([]ProfileImport, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.ListImports(ctx, profileID)
}
func (s *Service) GetImport(ctx context.Context, profileID, id uuid.UUID) (*ProfileImport, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.GetImport(ctx, profileID, id)
}
func (s *Service) ListImportCandidates(ctx context.Context, profileID, importID uuid.UUID) ([]ImportCandidateResponse, error) {
	if _, err := s.GetImport(ctx, profileID, importID); err != nil {
		return nil, err
	}
	rows, err := s.workflow.ListCandidates(ctx, importID)
	if err != nil {
		return nil, err
	}
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	out := make([]ImportCandidateResponse, 0, len(rows))
	for i := range rows {
		item := candidateResponse(&rows[i])
		if rows[i].MatchedEntityID != nil {
			item.Existing = findEffectiveEntity(effective, rows[i].SectionType, *rows[i].MatchedEntityID)
			item.Changes = fieldChanges(item.Existing, rows[i].CandidateData)
		}
		out = append(out, item)
	}
	return out, nil
}
func (s *Service) ReviewCandidate(ctx context.Context, profileID, importID, candidateID uuid.UUID, action string) (*ImportCandidateResponse, error) {
	imp, err := s.GetImport(ctx, profileID, importID)
	if err != nil {
		return nil, err
	}
	if imp.Status != "review_required" {
		return nil, ErrImportNotReady
	}
	candidate, err := s.workflow.GetCandidate(ctx, importID, candidateID)
	if err != nil {
		return nil, err
	}
	switch action {
	case "accept":
		candidate.Status = "accepted"
	case "reject":
		candidate.Status = "rejected"
	case "merge":
		if candidate.MatchedEntityID == nil {
			return nil, newValidationError("merge requires a matched existing entity")
		}
		candidate.Status = "merged"
	default:
		return nil, newValidationError("action must be accept, reject, or merge")
	}
	if err = s.workflow.UpdateCandidate(ctx, candidate); err != nil {
		return nil, err
	}
	review := candidateResponse(candidate)
	if candidate.MatchedEntityID != nil {
		effective, err := s.ResolveEffectiveProfile(ctx, profileID)
		if err != nil {
			return nil, err
		}
		review.Existing = findEffectiveEntity(effective, candidate.SectionType, *candidate.MatchedEntityID)
		review.Changes = fieldChanges(review.Existing, candidate.CandidateData)
	}
	return &review, nil
}

func (s *Service) ApplyImport(ctx context.Context, profileID, importID uuid.UUID) (*ProfileImport, error) {
	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	imp, err := s.workflow.GetImport(ctx, profileID, importID)
	if err != nil {
		return nil, err
	}
	if imp.Status != "review_required" {
		return nil, ErrImportNotReady
	}
	candidates, err := s.workflow.ListCandidates(ctx, importID)
	if err != nil {
		return nil, err
	}
	accepted := 0
	rejectedOrMerged := 0
	entities := []any{}
	evidence := []ProfileEvidence{}
	for i := range candidates {
		candidate := &candidates[i]
		switch candidate.Status {
		case "pending":
			return nil, ErrImportNotReady
		case "rejected", "merged":
			rejectedOrMerged++
			continue
		case "accepted":
			accepted++
		default:
			continue
		}
		entity, entityID, err := candidateEntity(profileID, candidate)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
		evidence = append(evidence, ProfileEvidence{ProfileID: profileID, EntityType: candidate.SectionType, EntityID: entityID, EvidenceType: "imported_cv", DocumentID: &imp.DocumentID, ImportID: &imp.ID, SourceText: candidate.SourceText, Confidence: candidate.Confidence})
	}
	if accepted == 0 {
		return nil, ErrImportNotReady
	}
	now := time.Now().UTC()
	imp.Status = "approved"
	if rejectedOrMerged > 0 {
		imp.Status = "partially_approved"
	}
	imp.CompletedAt = &now
	entityType := "profile_import"
	audit := &AuditLog{UserID: &profile.UserID, ProfileID: &profileID, Action: "profile.import.applied", EntityType: &entityType, EntityID: &imp.ID}
	if err = s.workflow.ApplyImport(ctx, imp, entities, evidence, audit); err != nil {
		return nil, err
	}
	return imp, nil
}

func (s *Service) buildCandidates(ctx context.Context, profileID, importID uuid.UUID, extracted *ExtractedProfile) ([]ProfileImportCandidate, error) {
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	rows := []ProfileImportCandidate{}
	add := func(section string, data any, confidence *float64, source *string) {
		encoded, _ := json.Marshal(data)
		candidate := ProfileImportCandidate{ImportID: importID, SectionType: section, CandidateData: JSON(encoded), SourceText: source, Confidence: confidence, Status: "pending"}
		candidate.MatchedEntityID = matchCandidate(effective, section, encoded)
		rows = append(rows, candidate)
	}
	if extracted.PersonalInfo != nil {
		add("personal_info", extracted.PersonalInfo.Data, extracted.PersonalInfo.Confidence, extracted.PersonalInfo.SourceText)
	}
	for _, x := range extracted.Education {
		add("education", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Employment {
		add("employment", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Projects {
		add("projects", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Publications {
		add("publications", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Articles {
		add("articles", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Skills {
		add("skills", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Certifications {
		add("certifications", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Awards {
		add("awards", x.Data, x.Confidence, x.SourceText)
	}
	for _, x := range extracted.Volunteering {
		add("volunteering", x.Data, x.Confidence, x.SourceText)
	}
	return rows, nil
}
func matchCandidate(effective *EffectiveProfileResponse, section string, data []byte) *uuid.UUID {
	var candidate map[string]any
	if json.Unmarshal(data, &candidate) != nil {
		return nil
	}
	if section == "personal_info" && effective.PersonalInfo != nil {
		id := effective.PersonalInfo.ID
		return &id
	}
	list, ok := sectionList(&effective.FullProfileResponse, section)
	if !ok {
		return nil
	}
	keys := []string{"name", "title", "institution", "organization"}
	for _, item := range *list {
		raw, _ := json.Marshal(item)
		var current map[string]any
		_ = json.Unmarshal(raw, &current)
		matched := false
		for _, key := range keys {
			left, leftOK := candidate[key].(string)
			right, rightOK := current[key].(string)
			if leftOK && rightOK && similarText(left, right) {
				matched = true
				break
			}
		}
		if matched {
			id := item.ID
			return &id
		}
	}
	return nil
}
func similarText(left, right string) bool {
	left = nonAlphaNumeric.ReplaceAllString(strings.ToLower(left), "")
	right = nonAlphaNumeric.ReplaceAllString(strings.ToLower(right), "")
	if left == "" || right == "" {
		return false
	}
	if left == right {
		return true
	}
	short, long := left, right
	if len(short) > len(long) {
		short, long = long, short
	}
	return len(short)*4 >= len(long)*3 && strings.Contains(long, short)
}
func findEffectiveEntity(effective *EffectiveProfileResponse, section string, id uuid.UUID) any {
	if section == "personal_info" && effective.PersonalInfo != nil && effective.PersonalInfo.ID == id {
		return effective.PersonalInfo
	}
	if list, ok := sectionList(&effective.FullProfileResponse, section); ok {
		for _, item := range *list {
			if item.ID == id {
				return item
			}
		}
	}
	return nil
}
func fieldChanges(existing any, candidateData []byte) []FieldChangeResponse {
	if existing == nil {
		return nil
	}
	raw, _ := json.Marshal(existing)
	var current, candidate map[string]any
	_ = json.Unmarshal(raw, &current)
	_ = json.Unmarshal(candidateData, &candidate)
	changes := []FieldChangeResponse{}
	for field, value := range candidate {
		if before, ok := current[field]; !ok || fmt.Sprint(before) != fmt.Sprint(value) {
			changes = append(changes, FieldChangeResponse{Field: field, Existing: before, Candidate: value})
		}
	}
	return changes
}
func candidateEntity(profileID uuid.UUID, candidate *ProfileImportCandidate) (any, uuid.UUID, error) {
	if candidate.SectionType == "personal_info" {
		var req struct {
			FirstName string  `json:"firstName"`
			LastName  string  `json:"lastName"`
			Phone     *string `json:"phone"`
		}
		if err := json.Unmarshal(candidate.CandidateData, &req); err != nil {
			return nil, uuid.Nil, err
		}
		if req.FirstName == "" || req.LastName == "" {
			return nil, uuid.Nil, newValidationError("personal information candidate is missing required names")
		}
		entity := &ProfilePersonalInfo{Base: Base{ID: uuid.New()}, ProfileID: profileID, FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone}
		return entity, entity.ID, nil
	}
	var req SectionRequest
	if err := json.Unmarshal(candidate.CandidateData, &req); err != nil {
		return nil, uuid.Nil, err
	}
	kind, ok := sectionKindFromEntityType(candidate.SectionType)
	if !ok {
		return nil, uuid.Nil, newValidationError("unsupported candidate section")
	}
	if err := validateSection(kind, req, true); err != nil {
		return nil, uuid.Nil, newValidationError(err.Error())
	}
	entity := newSection(kind, profileID, req)
	id := uuid.New()
	setEntityID(entity, id)
	return entity, id, nil
}
func sectionKindFromEntityType(value string) (SectionKind, bool) {
	switch normalizeEntityType(value) {
	case "education":
		return EducationKind, true
	case "employment":
		return EmploymentKind, true
	case "projects":
		return ProjectsKind, true
	case "publications":
		return PublicationsKind, true
	case "articles":
		return ArticlesKind, true
	case "skills":
		return SkillsKind, true
	case "certifications":
		return CertificationsKind, true
	case "awards":
		return AwardsKind, true
	case "volunteering":
		return VolunteeringKind, true
	}
	return "", false
}
func setEntityID(entity any, id uuid.UUID) {
	switch value := entity.(type) {
	case *EducationHistory:
		value.ID = id
	case *EmploymentHistory:
		value.ID = id
	case *Project:
		value.ID = id
	case *Publication:
		value.ID = id
	case *Article:
		value.ID = id
	case *Skill:
		value.ID = id
	case *ResearchInterest:
		value.ID = id
	case *CareerGoal:
		value.ID = id
	case *Certification:
		value.ID = id
	case *Award:
		value.ID = id
	case *VolunteerExperience:
		value.ID = id
	}
}
