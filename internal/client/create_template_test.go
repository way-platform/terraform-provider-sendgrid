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

func TestCreateTemplate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/templates" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("authorization = %q, want %q", got, "Bearer test-key")
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "My Template" || body["generation"] != "dynamic" {
			t.Errorf("unexpected body: %v", body)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id":         "d-abc123",
			"name":       "My Template",
			"generation": "dynamic",
			"updated_at": "2026-01-01T00:00:00Z",
		})
	}))
	defer srv.Close()

	gen := templates.GenerationDynamic
	c := client.New("test-key", client.WithBaseURL(srv.URL))
	tmpl, err := c.CreateTemplate(context.Background(), templates.CreateTemplateJSONRequestBody{
		Name:       "My Template",
		Generation: &gen,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &templates.TransactionalTemplate{
		ID:         "d-abc123",
		Name:       "My Template",
		Generation: templates.Generation1Dynamic,
		UpdatedAt:  "2026-01-01T00:00:00Z",
	}
	if diff := cmp.Diff(want, tmpl); diff != "" {
		t.Errorf("template mismatch (-want +got):\n%s", diff)
	}
}
