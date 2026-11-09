package workflow

import (
	"encoding/json"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"testing"
)

func TestLegacyProposedCatalogueName(t *testing.T) {
	raw := profile.JSON(`"The American University in Cairo"`)
	normalized, err := proposedRecordJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	var institution ProposedInstitution
	if err := json.Unmarshal(normalized, &institution); err != nil {
		t.Fatal(err)
	}
	if institution.Name != "The American University in Cairo" || institution.Country != "" || institution.WebsiteURL != nil {
		t.Fatalf("unsupported metadata inferred: %+v", institution)
	}
	if string(raw) != `"The American University in Cairo"` {
		t.Fatal("approved record modified")
	}
	structured := profile.JSON(`{"name":"University","country":"Egypt","websiteUrl":"https://example.edu"}`)
	normalized, err = proposedRecordJSON(structured)
	if err != nil || string(normalized) != string(structured) {
		t.Fatal("structured proposal changed", err)
	}
	if _, err := proposedRecordJSON(profile.JSON(`"   "`)); err == nil {
		t.Fatal("blank catalogue name accepted")
	}
	quoted, _ := json.Marshal(`University "Example"`)
	normalized, err = proposedRecordJSON(profile.JSON(quoted))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(normalized, &institution); err != nil || institution.Name != `University "Example"` {
		t.Fatal("escaped name changed", err)
	}
}
