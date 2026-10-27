package workflow

import (
	"encoding/json"
	"strings"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/go-playground/validator/v10"
)

// proposalFindingValue adapts research evidence to a supported application DTO.
// Research values are intentionally open-ended; a successful JSON decode alone
// does not mean they are safe or meaningful application records. Unsupported
// values stay linked to the research run and pending review.
func proposalFindingValue(finding *application.ResearchFinding, validate *validator.Validate) (json.RawMessage, bool) {
	if finding.ReviewStatus == "rejected" {
		return nil, false
	}
	var request any
	switch finding.Category {
	case "deadline":
		return proposalDeadlineValue(finding, validate)
	case "requirement":
		request = &application.RequirementRequest{}
	case "funding":
		request = &application.FundingRequest{}
	case "url":
		request = &application.URLRequest{}
	case "contact":
		request = &application.ContactRequest{}
	case "supervisor":
		request = &application.SupervisorRequest{}
	case "questionnaire":
		var candidate QuestionnaireCandidate
		if json.Unmarshal(finding.Value, &candidate) != nil || strings.TrimSpace(candidate.Title) == "" {
			return nil, false
		}
		for _, question := range candidate.Questions {
			if validate.Struct(question) != nil {
				return nil, false
			}
			if question.Status != nil {
				switch *question.Status {
				case "unanswered", "suggested", "answered", "needs_information", "needs_review", "approved":
				default:
					return nil, false
				}
			}
		}
		return json.RawMessage(finding.Value), true
	default:
		return nil, false
	}
	if json.Unmarshal(finding.Value, request) != nil || validate.Struct(request) != nil {
		return nil, false
	}
	if contact, ok := request.(*application.ContactRequest); ok {
		if (contact.Name == nil || strings.TrimSpace(*contact.Name) == "") && (contact.Email == nil || strings.TrimSpace(*contact.Email) == "") {
			return nil, false
		}
	}
	return json.RawMessage(finding.Value), true
}

func proposalDeadlineValue(finding *application.ResearchFinding, validate *validator.Validate) (json.RawMessage, bool) {
	var request application.DeadlineRequest
	if json.Unmarshal(finding.Value, &request) != nil {
		return nil, false
	}
	var aliases struct {
		RawText *string `json:"rawText"`
	}
	if json.Unmarshal(finding.Value, &aliases) != nil {
		return nil, false
	}
	if request.RawDeadlineText == nil || strings.TrimSpace(*request.RawDeadlineText) == "" {
		request.RawDeadlineText = aliases.RawText
	}
	if request.RawDeadlineText == nil || strings.TrimSpace(*request.RawDeadlineText) == "" {
		request.RawDeadlineText = finding.RawText
	}
	if request.DeadlineAt == nil && (request.RawDeadlineText == nil || strings.TrimSpace(*request.RawDeadlineText) == "") {
		return nil, false
	}
	if strings.TrimSpace(request.Title) == "" {
		request.Title = finding.Field
	}
	if strings.TrimSpace(request.DeadlineType) == "" {
		request.DeadlineType = "other"
	}
	switch request.DatePrecision {
	case "date":
		request.DatePrecision = "day"
	case "exact", "day", "month", "approximate", "unknown":
	default:
		request.DatePrecision = "unknown"
	}
	// Never derive a timestamp from localDate, recurring dates or vague prose.
	// An approximate/month/unknown deadline cannot justify an exact instant.
	if request.DatePrecision == "approximate" || request.DatePrecision == "month" || request.DatePrecision == "unknown" {
		request.DeadlineAt = nil
	}
	if request.DatePrecision == "exact" && request.DeadlineAt == nil {
		request.DatePrecision = "unknown"
	}
	if request.DeadlineAt == nil && (request.RawDeadlineText == nil || strings.TrimSpace(*request.RawDeadlineText) == "") {
		return nil, false
	}
	if request.Notes == nil {
		request.Notes = finding.Notes
	}
	if validate.Struct(request) != nil {
		return nil, false
	}
	value, err := json.Marshal(request)
	return value, err == nil
}
