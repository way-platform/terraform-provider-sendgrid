package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestDeleteImage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/images/img-001" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	if err := c.DeleteImage(context.Background(), "img-001"); err != nil {
		t.Fatal(err)
	}
}
