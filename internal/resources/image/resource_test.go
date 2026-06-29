package image_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/way-platform/terraform-provider-sendgrid/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sendgrid": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccImage(t *testing.T) {
	// Create a temp image file for upload.
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "logo.png")
	// 1x1 transparent PNG.
	png1x1 := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x01, 0xe5, 0x27, 0xde, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
		0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(imgPath, png1x1, 0o600); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	images := map[string]map[string]any{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/images":
			// Parse multipart to validate the upload field exists.
			//nolint:gosec // G120: bounded 10MB limit in a test stub handler
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]any{
					"errors": []map[string]string{{"message": "bad multipart"}},
				})
				return
			}
			img := map[string]any{
				"id":     "img-001",
				"name":   "logo.png",
				"url":    "https://cdn.sendgrid.example/img-001.png",
				"width":  1,
				"height": 1,
			}
			images["img-001"] = img
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(img)

		case r.Method == http.MethodGet && r.URL.Path == "/v3/images/img-001":
			img, ok := images["img-001"]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{
					"errors": []map[string]string{{"message": "not found"}},
				})
				return
			}
			json.NewEncoder(w).Encode(img)

		case r.Method == http.MethodDelete && r.URL.Path == "/v3/images/img-001":
			delete(images, "img-001")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"errors": []map[string]string{{"message": "not found"}},
			})
		}
	}))
	defer srv.Close()

	t.Setenv("SENDGRID_API_KEY", "test-key")
	t.Setenv("SENDGRID_API_URL", srv.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig(imgPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_image.test", "id", "img-001"),
					resource.TestCheckResourceAttr("sendgrid_image.test", "name", "logo.png"),
					resource.TestCheckResourceAttr(
						"sendgrid_image.test",
						"url",
						"https://cdn.sendgrid.example/img-001.png",
					),
					resource.TestCheckResourceAttr("sendgrid_image.test", "width", "1"),
					resource.TestCheckResourceAttr("sendgrid_image.test", "height", "1"),
				),
			},
			{
				ResourceName:            "sendgrid_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"file_path", "file_sha256"},
			},
		},
	})
}

func testAccImageConfig(path string) string {
	return `
		resource "sendgrid_image" "test" {
			file_path   = "` + path + `"
			file_sha256 = filesha256("` + path + `")
		}
	`
}
