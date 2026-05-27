package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestDeleteTemplateVersion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/templates/d-abc123/versions/ver-001" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	if err := c.DeleteTemplateVersion(context.Background(), "d-abc123", "ver-001"); err != nil {
		t.Fatal(err)
	}
}
