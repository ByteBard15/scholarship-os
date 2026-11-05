package writing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWritingPaginationRejectsInvalidParameters(t *testing.T) {
	// Invalid pagination must be rejected before accessing the service/database.
	handler := &Handler{}
	for _, query := range []string{"page=0", "page=-1", "page=abc", "page=1000001", "pageSize=0", "pageSize=101", "pageSize=1.5"} {
		t.Run(query, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ListSamples(recorder, httptest.NewRequest(http.MethodGet, "/writing-samples?"+query, nil))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", recorder.Code)
			}
		})
	}
}
