package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{
				{"message": "resource not found", "field": "id"},
			},
		})
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	_, err := c.GetTemplate(context.Background(), "d-missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected not found, got: %v", err)
	}
}
