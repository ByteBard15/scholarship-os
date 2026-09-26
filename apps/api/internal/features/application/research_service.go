package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	appLogger "github.com/example/scholarship-os/apps/api/pkg/logger"
	"github.com/google/uuid"
)

type ResearchService struct {
	repo         ResearchRepository
	applications *Service
	researcher   ApplicationResearcher
	log          *slog.Logger
	now          func() time.Time
}

func NewResearchService(r ResearchRepository, a *Service, p ApplicationResearcher, l *slog.Logger) *ResearchService {
	return &ResearchService{r, a, p, l, func() time.Time { return time.Now().UTC() }}
}
func (s *ResearchService) Run(ctx context.Context, applicationID uuid.UUID, q ResearchRunRequest) (*ResearchRun, error) {
	app, e := s.applications.repo.Get(ctx, applicationID)
	if e != nil {
		return nil, e
	}
	if q.Trigger == "" {
		q.Trigger = "manual"
	}
	if q.ResearchType == "" {
		q.ResearchType = "full"
	}
	now := s.now()
	provider, model := s.researcher.ProviderName(), s.researcher.ModelName()
	prompt := "phase3-v1"
	run := &ResearchRun{ApplicationID: &applicationID, Status: "running", Trigger: q.Trigger, ResearchType: q.ResearchType, ModelProvider: &provider, ModelName: &model, PromptVersion: &prompt, StartedAt: &now}
	if app.Status == StatusDiscovered {
		app.Status = StatusResearching
	}
	app.ResearchStatus = ResearchRunning
	if e = s.repo.CreateRun(ctx, run, app); e != nil {
		return nil, e
	}
	appLogger.FromContext(ctx, s.log).InfoContext(ctx, "research started", "application_id", applicationID, "run_id", run.ID, "provider", provider)
	effective, e := s.applications.profiles.ResolveEffectiveProfile(ctx, app.ApplicantProfileID)
	if e != nil {
		_ = s.repo.FailRun(ctx, run, app, e)
		return nil, e
	}
	summary := fmt.Sprintf("profile %s with %d education, %d employment, and %d skill entries", effective.Name, len(effective.Education), len(effective.Employment), len(effective.Skills))
	result, e := s.researcher.Research(ctx, ResearchInput{Application: toApplication(app), EffectiveProfileSummary: summary})
	if e != nil {
		_ = s.repo.FailRun(ctx, run, app, e)
		appLogger.FromContext(ctx, s.log).ErrorContext(ctx, "research failed", "application_id", applicationID, "run_id", run.ID, "error", e)
		return nil, e
	}
	sources, findings, e := buildResearchRecords(run, app, result)
	if e != nil {
		_ = s.repo.FailRun(ctx, run, app, e)
		return nil, e
	}
	completed := s.now()
	run.Status = "review_required"
	run.CompletedAt = &completed
	app.ResearchStatus = ResearchReviewRequired
	if e = s.repo.PersistResult(ctx, run, app, sources, findings); e != nil {
		return nil, e
	}
	appLogger.FromContext(ctx, s.log).InfoContext(ctx, "research completed", "application_id", applicationID, "run_id", run.ID, "sources", len(sources), "findings", len(findings))
	return run, nil
}

func buildResearchRecords(run *ResearchRun, app *Application, result *ResearchResult) ([]ResearchSource, []ResearchFinding, error) {
	sourceIDs := map[string]uuid.UUID{}
	sources := make([]ResearchSource, len(result.Sources))
	for i, v := range result.Sources {
		parsed, err := url.Parse(v.URL)
		if v.Key == "" || err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, nil, newValidationError("research source requires a unique key and valid HTTP(S) URL")
		}
		if _, exists := sourceIDs[v.Key]; exists {
			return nil, nil, newValidationError("research source keys must be unique")
		}
		if v.RetrievedAt.IsZero() {
			return nil, nil, newValidationError("research source retrieved_at is required")
		}
		id := uuid.New()
		sourceIDs[v.Key] = id
		sources[i] = ResearchSource{ID: id, ResearchRunID: run.ID, ApplicationID: &app.ID, URL: v.URL, Title: v.Title, Publisher: v.Publisher, SourceType: v.SourceType, IsOfficial: v.IsOfficial, RetrievedAt: v.RetrievedAt, PublishedAt: v.PublishedAt, ContentHash: v.ContentHash, Notes: v.Notes}
	}
	findings := make([]ResearchFinding, 0, len(result.Requirements)+len(result.Deadlines)+len(result.Funding)+len(result.Contacts)+len(result.Supervisors)+len(result.URLs)+len(result.Findings))
	add := func(sourceKey, category, field string, value any, confidence *float64, raw *string, verification string, notes *string) error {
		if strings.TrimSpace(category) == "" || strings.TrimSpace(field) == "" {
			return newValidationError("research findings require category and field")
		}
		if sourceKey != "" {
			if _, ok := sourceIDs[sourceKey]; !ok {
				return newValidationError("research finding references an unknown source")
			}
		}
		payload, e := json.Marshal(value)
		if e != nil {
			return e
		}
		var sourceID *uuid.UUID
		if id, ok := sourceIDs[sourceKey]; ok {
			x := id
			sourceID = &x
		}
		if verification == "" {
			verification = "unverified"
		}
		if !validVerificationStatus(verification) {
			return newValidationError("research finding verification status is invalid")
		}
		findings = append(findings, ResearchFinding{ID: uuid.New(), ResearchRunID: run.ID, ApplicationID: &app.ID, SourceID: sourceID, Category: category, Field: field, Value: profile.JSON(payload), Confidence: confidence, VerificationStatus: verification, ReviewStatus: "pending", RawText: raw, Notes: notes})
		return nil
	}
	for _, v := range result.Requirements {
		if e := add(v.SourceKey, "requirement", v.Data.Title, v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.Deadlines {
		if e := add(v.SourceKey, "deadline", v.Data.DeadlineType, v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.Funding {
		if e := add(v.SourceKey, "funding", v.Data.FundingType, v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.Contacts {
		if e := add(v.SourceKey, "contact", valueOr(v.Data.ContactType, "contact"), v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.Supervisors {
		if e := add(v.SourceKey, "supervisor", v.Data.Name, v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.URLs {
		if e := add(v.SourceKey, "url", v.Data.URLType, v.Data, v.Confidence, v.RawText, v.VerificationStatus, nil); e != nil {
			return nil, nil, e
		}
	}
	for _, v := range result.Findings {
		if e := add(v.SourceKey, v.Category, v.Field, v.Value, v.Confidence, v.RawText, v.VerificationStatus, v.Notes); e != nil {
			return nil, nil, e
		}
	}
	for i := range findings {
		for j := i + 1; j < len(findings); j++ {
			if findings[i].Category == findings[j].Category && findings[i].Field == findings[j].Field && string(findings[i].Value) != string(findings[j].Value) {
				findings[i].VerificationStatus = "conflicting"
				findings[j].VerificationStatus = "conflicting"
			}
		}
	}
	return sources, findings, nil
}
func valueOr(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	return *v
}
func (s *ResearchService) ListRuns(c context.Context, a uuid.UUID) ([]ResearchRun, error) {
	if _, e := s.applications.repo.Get(c, a); e != nil {
		return nil, e
	}
	return s.repo.ListRuns(c, a)
}
func (s *ResearchService) GetRun(c context.Context, a, id uuid.UUID) (*ResearchRun, error) {
	return s.repo.GetRun(c, a, id)
}
func (s *ResearchService) ListSources(c context.Context, a, run uuid.UUID) ([]ResearchSource, error) {
	if _, e := s.repo.GetRun(c, a, run); e != nil {
		return nil, e
	}
	return s.repo.ListSources(c, run)
}
func (s *ResearchService) AddSource(c context.Context, a, runID uuid.UUID, q ResearchSourceCandidate) (*ResearchSource, error) {
	run, e := s.repo.GetRun(c, a, runID)
	if e != nil {
		return nil, e
	}
	if run.Status == "complete" || run.Status == "failed" {
		return nil, ErrInvalidResearchState
	}
	parsed, e := url.Parse(q.URL)
	if e != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, newValidationError("research source requires a valid HTTP(S) URL")
	}
	if q.RetrievedAt.IsZero() {
		q.RetrievedAt = s.now()
	}
	v := &ResearchSource{ResearchRunID: runID, ApplicationID: &a, URL: q.URL, Title: q.Title, Publisher: q.Publisher, SourceType: q.SourceType, IsOfficial: q.IsOfficial, RetrievedAt: q.RetrievedAt, PublishedAt: q.PublishedAt, ContentHash: q.ContentHash, Notes: q.Notes}
	if strings.TrimSpace(v.SourceType) == "" {
		return nil, newValidationError("research source type is required")
	}
	return v, s.repo.CreateSource(c, v)
}
func (s *ResearchService) ListFindings(c context.Context, a, run uuid.UUID) ([]ResearchFinding, error) {
	if _, e := s.repo.GetRun(c, a, run); e != nil {
		return nil, e
	}
	return s.repo.ListFindings(c, run)
}
func (s *ResearchService) AddFinding(c context.Context, a, runID uuid.UUID, q FindingProposal) (*ResearchFinding, error) {
	run, e := s.repo.GetRun(c, a, runID)
	if e != nil {
		return nil, e
	}
	if run.Status == "complete" || run.Status == "failed" {
		return nil, ErrInvalidResearchState
	}
	if strings.TrimSpace(q.Category) == "" || strings.TrimSpace(q.Field) == "" {
		return nil, newValidationError("research findings require category and field")
	}
	if q.Confidence != nil && (*q.Confidence < 0 || *q.Confidence > 1) {
		return nil, newValidationError("research finding confidence must be between 0 and 1")
	}
	if q.SourceID != nil {
		sources, err := s.repo.ListSources(c, runID)
		if err != nil {
			return nil, err
		}
		found := false
		for _, source := range sources {
			if source.ID == *q.SourceID {
				found = true
				break
			}
		}
		if !found {
			return nil, newValidationError("research finding source does not belong to this run")
		}
	}
	payload, e := json.Marshal(q.Value)
	if e != nil {
		return nil, e
	}
	verification := q.VerificationStatus
	if verification == "" {
		verification = "unverified"
	}
	if !validVerificationStatus(verification) {
		return nil, newValidationError("research finding verification status is invalid")
	}
	v := &ResearchFinding{ResearchRunID: runID, ApplicationID: &a, SourceID: q.SourceID, Category: q.Category, Field: q.Field, Value: profile.JSON(payload), Confidence: q.Confidence, VerificationStatus: verification, ReviewStatus: "pending", RawText: q.RawText, Notes: q.Notes}
	return v, s.repo.CreateFinding(c, v)
}
func (s *ResearchService) ReviewFinding(c context.Context, a, run, id uuid.UUID, status string) (*ResearchFinding, error) {
	if _, e := s.repo.GetRun(c, a, run); e != nil {
		return nil, e
	}
	v, e := s.repo.GetFinding(c, run, id)
	if e != nil {
		return nil, e
	}
	v.ReviewStatus = status
	return v, s.repo.UpdateFinding(c, v)
}
func (s *ResearchService) Apply(c context.Context, a, runID uuid.UUID) error {
	run, e := s.repo.GetRun(c, a, runID)
	if e != nil {
		return e
	}
	if run.Status != "review_required" {
		return ErrInvalidResearchState
	}
	app, e := s.applications.repo.Get(c, a)
	if e != nil {
		return e
	}
	findings, e := s.repo.ListFindings(c, runID)
	if e != nil {
		return e
	}
	batch := AppliedResearch{}
	accepted := 0
	for _, f := range findings {
		if f.ReviewStatus == "pending" {
			return ErrResearchNotReady
		}
		if f.ReviewStatus == "rejected" {
			continue
		}
		if f.ReviewStatus != "accepted" && f.ReviewStatus != "auto_accepted" {
			continue
		}
		accepted++
		switch f.Category {
		case "requirement":
			var q RequirementRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			if strings.TrimSpace(q.Category) == "" || strings.TrimSpace(q.Title) == "" {
				return newValidationError("accepted requirement finding is missing category or title")
			}
			mandatory := true
			if q.IsMandatory != nil {
				mandatory = *q.IsMandatory
			}
			status := "unknown"
			if q.Status != nil {
				status = *q.Status
			}
			ev := false
			if q.EvidenceRequired != nil {
				ev = *q.EvidenceRequired
			}
			batch.Requirements = append(batch.Requirements, ApplicationRequirement{ApplicationID: a, Category: q.Category, Title: q.Title, Description: q.Description, IsMandatory: mandatory, Status: status, DueDate: q.DueDate, SourceID: f.SourceID, SourceURL: q.SourceURL, EvidenceRequired: ev, Notes: q.Notes, SortOrder: q.SortOrder})
		case "deadline":
			var q DeadlineRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			if strings.TrimSpace(q.DeadlineType) == "" || strings.TrimSpace(q.Title) == "" || !validDatePrecision(q.DatePrecision) {
				return newValidationError("accepted deadline finding is missing required fields")
			}
			if q.DeadlineAt == nil && (q.DatePrecision == "exact" || q.DatePrecision == "day") {
				return newValidationError("accepted exact deadline finding requires deadline_at")
			}
			hard := false
			if q.IsHardDeadline != nil {
				hard = *q.IsHardDeadline
			}
			batch.Deadlines = append(batch.Deadlines, ApplicationDeadline{ApplicationID: a, DeadlineType: q.DeadlineType, Title: q.Title, DeadlineAt: q.DeadlineAt, Timezone: q.Timezone, DatePrecision: q.DatePrecision, RawDeadlineText: q.RawDeadlineText, IsHardDeadline: hard, SourceID: f.SourceID, SourceURL: q.SourceURL, VerifiedAt: q.VerifiedAt, Notes: q.Notes})
		case "funding":
			var q FundingRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			if strings.TrimSpace(q.FundingType) == "" {
				return newValidationError("accepted funding finding is missing funding_type")
			}
			v := fundingFrom(a, q)
			v.SourceID = f.SourceID
			batch.Funding = append(batch.Funding, *v)
		case "contact":
			var q ContactRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			v := contactFrom(a, q)
			v.SourceID = f.SourceID
			batch.Contacts = append(batch.Contacts, *v)
		case "supervisor":
			var q SupervisorRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			if strings.TrimSpace(q.Name) == "" {
				return newValidationError("accepted supervisor finding is missing name")
			}
			v := supervisorFrom(a, q)
			v.SourceID = f.SourceID
			batch.Supervisors = append(batch.Supervisors, *v)
		case "url":
			var q URLRequest
			if e = json.Unmarshal(f.Value, &q); e != nil {
				return e
			}
			parsed, parseErr := url.Parse(q.URL)
			if strings.TrimSpace(q.URLType) == "" || parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return newValidationError("accepted URL finding is invalid")
			}
			batch.URLs = append(batch.URLs, ApplicationURL{ApplicationID: a, URLType: q.URLType, Label: q.Label, URL: q.URL, IsOfficial: q.IsOfficial})
		}
	}
	if accepted == 0 {
		return ErrResearchNotReady
	}
	audit := s.applications.audit(app, "research.findings.applied", "research_run", run.ID)
	meta, _ := json.Marshal(map[string]any{"accepted_findings": accepted})
	audit.Metadata = profile.JSON(meta)
	if e = s.repo.ApplyFindings(c, run, app, batch, audit); e != nil {
		return e
	}
	appLogger.FromContext(c, s.log).InfoContext(c, "research findings applied", "application_id", a, "run_id", runID, "count", accepted)
	return nil
}

func validDatePrecision(v string) bool {
	return v == "exact" || v == "day" || v == "month" || v == "approximate" || v == "unknown"
}
func validVerificationStatus(v string) bool {
	return v == "verified" || v == "supported" || v == "conflicting" || v == "unverified" || v == "inferred"
}
func (s *ResearchService) MarkStale(c context.Context, a uuid.UUID) error {
	app, e := s.applications.repo.Get(c, a)
	if e != nil {
		return e
	}
	app.ResearchStatus = ResearchStale
	return s.repo.MarkStale(c, app)
}
