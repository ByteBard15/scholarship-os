package workflow

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
)

// Early research proposals stored catalogue names as JSON strings. Interpret
// those approved names without rewriting the proposal or fabricating metadata.
func proposedRecordJSON(raw profile.JSON) ([]byte, error) {
	data := bytes.TrimSpace(raw)
	if len(data) == 0 || data[0] != '"' {
		return data, nil
	}
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, validationError("approved proposal contains an empty catalogue name")
	}
	return json.Marshal(struct {
		Name string `json:"name"`
	}{Name: name})
}
