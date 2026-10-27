package workflow

import (
	"encoding/json"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/go-playground/validator/v10"
)

func TestProposalDeadlinePreservesPrecisionAndRawEvidence(t *testing.T) {
	tests := []struct {
		name, value, precision string
		timestamp              bool
	}{
		{"date-only research", `{"rawText":"8 June 2027; international applicants only","datePrecision":"date","localDate":"2027-06-08"}`, "day", false},
		{"exact", `{"rawText":"28 February 2027,17:00 ET","datePrecision":"exact","deadlineAt":"2027-02-28T22:00:00Z"}`, "exact", true},
		{"conditional", `{"rawText":"Before the offer deadline, if a new round opens","datePrecision":"conditional"}`, "unknown", false},
		{"recurring", `{"rawText":"July 7 annually; 2027 call unconfirmed","datePrecision":"recurring"}`, "unknown", false},
		{"multiple", `{"rawText":"November for nomination; January for admission","datePrecision":"multiple"}`, "unknown", false},
		{"month", `{"rawText":"June 2027","datePrecision":"month","deadlineAt":"2027-06-01T00:00:00Z"}`, "month", false},
		{"approximate", `{"rawText":"Around spring 2027","datePrecision":"approximate","deadlineAt":"2027-04-01T00:00:00Z"}`, "approximate", false},
		{"missing precision", `{"rawDeadlineText":"Ask the programme for the closing date"}`, "unknown", false},
		{"exact without time", `{"rawText":"Time not confirmed","datePrecision":"exact"}`, "unknown", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			finding := application.ResearchFinding{Category: "deadline", Field: "internationalAdmissionDeadline", Value: profile.JSON(test.value), ReviewStatus: "pending"}
			value, ok := proposalFindingValue(&finding, validator.New())
			if !ok {
				t.Fatal("deadline lost")
			}
			var got application.DeadlineRequest
			if err := json.Unmarshal(value, &got); err != nil {
				t.Fatal(err)
			}
			if got.DatePrecision != test.precision || (got.DeadlineAt != nil) != test.timestamp || got.RawDeadlineText == nil || got.Title != finding.Field || got.DeadlineType != "other" {
				t.Fatalf("bad adaptation: %+v", got)
			}
			if string(finding.Value) != test.value || finding.ReviewStatus != "pending" {
				t.Fatal("original research was mutated")
			}
		})
	}
}

func TestProposalFindingsLeaveUnsupportedValuesForReview(t *testing.T) {
	tests := []struct{ category, value, status string }{
		{"deadline", `{"datePrecision":"date","localDate":"2027-06-08"}`, "pending"},
		{"deadline", `{"rawText":"2027","deadlineAt":"not-a-timestamp"}`, "pending"},
		{"requirement", `{"entry":"Bachelor degree; university assessment required"}`, "pending"},
		{"requirement", `{"title":"Degree","category":"academic","status":"eligible"}`, "pending"},
		{"funding", `{"currency":"CAD","amount":44500}`, "pending"},
		{"funding", `{"fundingType":"stipend","amount":-100}`, "pending"},
		{"url", `{"url":"not a URL","urlType":"programme"}`, "pending"},
		{"contact", `{"role":"Potential supervisor"}`, "pending"},
		{"supervisor", `{"name":"Professor","contactStatus":"confirmed"}`, "pending"},
		{"questionnaire", `{"title":"Research interests","questions":[{"prompt":"Topic?","questionType":"text","status":"invalid"}]}`, "pending"},
		{"requirement", `{"title":"Degree","category":"academic"}`, "rejected"},
		{"general", `{"fit":"Strong"}`, "pending"},
		{"deadline", `null`, "pending"},
	}
	for _, test := range tests {
		t.Run(test.category+test.value, func(t *testing.T) {
			f := application.ResearchFinding{Category: test.category, Field: "field", Value: profile.JSON(test.value), ReviewStatus: test.status}
			if _, ok := proposalFindingValue(&f, validator.New()); ok {
				t.Fatal("unsupported evidence promoted into application")
			}
		})
	}
}

func TestProposalFindingsAcceptValidTypedRecords(t *testing.T) {
	tests := map[string]string{
		"requirement":   `{"category":"academic","title":"Degree","status":"unknown"}`,
		"funding":       `{"fundingType":"stipend","currency":"CAD","amount":44500}`,
		"url":           `{"urlType":"programme","url":"https://example.edu/programme"}`,
		"contact":       `{"email":"professor@example.edu"}`,
		"supervisor":    `{"name":"Professor","contactStatus":"not_contacted"}`,
		"questionnaire": `{"title":"Research interests","questions":[{"prompt":"Topic?","questionType":"text"}]}`,
	}
	for category, value := range tests {
		t.Run(category, func(t *testing.T) {
			f := application.ResearchFinding{Category: category, Value: profile.JSON(value), ReviewStatus: "pending"}
			if _, ok := proposalFindingValue(&f, validator.New()); !ok {
				t.Fatal("valid record discarded")
			}
		})
	}
}
