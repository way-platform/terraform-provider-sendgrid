package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestCreateTemplateVersion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/templates/d-abc123/versions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body templates.CreateTemplateVersionJSONRequestBody
		json.NewDecoder(r.Body).Decode(&body)
		if body.Subject != "Hello {{name}}" {
			t.Errorf("unexpected subject: %q", body.Subject)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":           "ver-001",
			"template_id":  "d-abc123",
			"name":         "v1",
			"subject":      body.Subject,
			"html_content": *(body.HTMLContent),
			"active":       1,
			"editor":       "code",
			"updated_at":   "2026-01-01T00:00:00Z",
		})
	}))
	defer srv.Close()

	active := templates.Active(1)
	editor := templates.EditorCode
	c := client.New("test-key", client.WithBaseURL(srv.URL))
	tv, err := c.CreateTemplateVersion(context.Background(), "d-abc123", templates.CreateTemplateVersionJSONRequestBody{
		Name:        "v1",
		Subject:     "Hello {{name}}",
		HTMLContent: client.Ptr("<h1>Hi</h1>"),
		Active:      &active,
		Editor:      &editor,
	})
	if err != nil {
		t.Fatal(err)
	}
	if *(tv.ID) != "ver-001" || *(tv.TemplateID) != "d-abc123" {
		t.Errorf("unexpected version: %+v", tv)
	}
}
