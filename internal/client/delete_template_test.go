package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestDeleteTemplate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/templates/d-abc123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	if err := c.DeleteTemplate(context.Background(), "d-abc123"); err != nil {
		t.Fatal(err)
	}
}
