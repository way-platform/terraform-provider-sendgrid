package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

func TestUploadImage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/images" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		ct := r.Header.Get("Content-Type")
		if len(ct) < 19 || ct[:19] != "multipart/form-data" {
			t.Errorf("unexpected content-type: %q", ct)
		}

		file, header, err := r.FormFile("upload")
		if err != nil {
			t.Fatalf("reading form file: %v", err)
		}
		defer file.Close()

		if header.Filename != "test.png" {
			t.Errorf("filename = %q, want %q", header.Filename, "test.png")
		}
		data, _ := io.ReadAll(file)
		if string(data) != "fake-png-data" {
			t.Errorf("unexpected file content: %q", data)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"id":     "img-001",
			"name":   "test.png",
			"url":    "https://cdn.sendgrid.net/img-001/100x100.png",
			"width":  100,
			"height": 100,
		})
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(testFile, []byte("fake-png-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := client.New("test-key", client.WithBaseURL(srv.URL))
	img, err := c.UploadImage(context.Background(), testFile)
	if err != nil {
		t.Fatal(err)
	}

	want := &client.Image{
		ID:     "img-001",
		Name:   "test.png",
		URL:    "https://cdn.sendgrid.net/img-001/100x100.png",
		Width:  100,
		Height: 100,
	}
	if diff := cmp.Diff(want, img); diff != "" {
		t.Errorf("image mismatch (-want +got):\n%s", diff)
	}
}
