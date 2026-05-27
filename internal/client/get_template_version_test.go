package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestGetTemplateVersion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/templates/d-abc123/versions/ver-001" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":           "ver-001",
			"template_id":  "d-abc123",
			"name":         "v1",
			"subject":      "Hello {{name}}",
			"html_content": "<h1>Hi</h1>",
			"active":       1,
			"editor":       "code",
			"updated_at":   "2026-01-01T00:00:00Z",
		})
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	tv, err := c.GetTemplateVersion(context.Background(), "d-abc123", "ver-001")
	if err != nil {
		t.Fatal(err)
	}

	active := templates.Active1(1)
	editor := templates.Editor1Code
	want := &templates.TransactionalTemplateVersionOutput{
		ID:          client.Ptr("ver-001"),
		TemplateID:  client.Ptr("d-abc123"),
		Name:        "v1",
		Subject:     "Hello {{name}}",
		HTMLContent: client.Ptr("<h1>Hi</h1>"),
		Active:      &active,
		Editor:      &editor,
		UpdatedAt:   client.Ptr("2026-01-01T00:00:00Z"),
	}
	if diff := cmp.Diff(want, tv); diff != "" {
		t.Errorf("template version mismatch (-want +got):\n%s", diff)
	}
}
