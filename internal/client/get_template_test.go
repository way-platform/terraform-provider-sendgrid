package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestGetTemplate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/templates/d-abc123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":         "d-abc123",
			"name":       "My Template",
			"generation": "dynamic",
			"updated_at": "2026-01-01T00:00:00Z",
			"versions":   []any{},
		})
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	tmpl, err := c.GetTemplate(context.Background(), "d-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.ID != "d-abc123" {
		t.Errorf("got ID %q, want %q", tmpl.ID, "d-abc123")
	}
}
