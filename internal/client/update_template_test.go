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

func TestUpdateTemplate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v3/templates/d-abc123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Updated Name" {
			t.Errorf("unexpected name: %q", body["name"])
		}

		json.NewEncoder(w).Encode(map[string]string{
			"id":         "d-abc123",
			"name":       "Updated Name",
			"generation": "dynamic",
			"updated_at": "2026-01-02T00:00:00Z",
		})
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	tmpl, err := c.UpdateTemplate(context.Background(), "d-abc123", templates.UpdateTemplateJSONRequestBody{
		Name: client.Ptr("Updated Name"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &templates.TransactionalTemplate{
		ID:         "d-abc123",
		Name:       "Updated Name",
		Generation: templates.Generation1Dynamic,
		UpdatedAt:  "2026-01-02T00:00:00Z",
	}
	if diff := cmp.Diff(want, tmpl); diff != "" {
		t.Errorf("template mismatch (-want +got):\n%s", diff)
	}
}
