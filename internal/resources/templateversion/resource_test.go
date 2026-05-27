package templateversion_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAccTemplateVersion(t *testing.T) {
	var mu sync.Mutex
	templates := map[string]map[string]any{
		"d-tmpl-001": {
			"id":         "d-tmpl-001",
			"name":       "existing",
			"generation": "dynamic",
			"updated_at": "2026-01-01 00:00:00",
		},
	}
	versions := map[string]map[string]any{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		// Template endpoints (needed for the template resource in config)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/templates":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			tmpl := map[string]any{
				"id":         "d-tmpl-001",
				"name":       body["name"],
				"generation": body["generation"],
				"updated_at": "2026-01-01 00:00:00",
			}
			templates["d-tmpl-001"] = tmpl
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(tmpl)

		case r.Method == http.MethodGet && r.URL.Path == "/v3/templates/d-tmpl-001":
			json.NewEncoder(w).Encode(templates["d-tmpl-001"])

		case r.Method == http.MethodDelete && r.URL.Path == "/v3/templates/d-tmpl-001":
			delete(templates, "d-tmpl-001")
			w.WriteHeader(http.StatusNoContent)

		// Template version endpoints
		case r.Method == http.MethodPost && r.URL.Path == "/v3/templates/d-tmpl-001/versions":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			ver := map[string]any{
				"id":           "ver-001",
				"template_id":  "d-tmpl-001",
				"name":         body["name"],
				"subject":      body["subject"],
				"html_content": body["html_content"],
				"active":       body["active"],
				"editor":       body["editor"],
				"updated_at":   "2026-01-01 00:00:00",
			}
			versions["ver-001"] = ver
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ver)

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v3/templates/d-tmpl-001/versions/"):
			vID := strings.TrimPrefix(r.URL.Path, "/v3/templates/d-tmpl-001/versions/")
			ver, ok := versions[vID]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{
					"errors": []map[string]string{{"message": "not found"}},
				})
				return
			}
			json.NewEncoder(w).Encode(ver)

		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v3/templates/d-tmpl-001/versions/"):
			vID := strings.TrimPrefix(r.URL.Path, "/v3/templates/d-tmpl-001/versions/")
			delete(versions, vID)
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
				Config: `
					resource "sendgrid_template" "test" {
						name = "version-test"
					}

					resource "sendgrid_template_version" "test" {
						template_id  = sendgrid_template.test.id
						name         = "v1"
						subject      = "Hello {{name}}"
						html_content = "<h1>Hi</h1>"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "id", "ver-001"),
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "template_id", "d-tmpl-001"),
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "subject", "Hello {{name}}"),
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "active", "1"),
				),
			},
			{
				ResourceName:      "sendgrid_template_version.test",
				ImportState:       true,
				ImportStateId:     "d-tmpl-001/ver-001",
				ImportStateVerify: true,
			},
		},
	})
}
