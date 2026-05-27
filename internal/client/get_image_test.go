package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestGetImage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/images/img-001" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":     "img-001",
			"name":   "logo.png",
			"url":    "https://cdn.sendgrid.net/img-001/200x50.png",
			"width":  200,
			"height": 50,
		})
	}))
	defer srv.Close()

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	img, err := c.GetImage(context.Background(), "img-001")
	if err != nil {
		t.Fatal(err)
	}

	want := &client.Image{
		ID:     "img-001",
		Name:   "logo.png",
		URL:    "https://cdn.sendgrid.net/img-001/200x50.png",
		Width:  200,
		Height: 50,
	}
	if diff := cmp.Diff(want, img); diff != "" {
		t.Errorf("image mismatch (-want +got):\n%s", diff)
	}
}
