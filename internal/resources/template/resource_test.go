package template_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestAccTemplate(t *testing.T) {
	var mu sync.Mutex
	templates := map[string]map[string]string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/templates":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			tmpl := map[string]string{
				"id":         "d-test-123",
				"name":       body["name"],
				"generation": body["generation"],
				"updated_at": "2026-01-01 00:00:00",
			}
			templates["d-test-123"] = tmpl
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(tmpl)

		case r.Method == http.MethodGet && r.URL.Path == "/v3/templates/d-test-123":
			tmpl, ok := templates["d-test-123"]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{
					"errors": []map[string]string{{"message": "not found"}},
				})
				return
			}
			json.NewEncoder(w).Encode(tmpl)

		case r.Method == http.MethodPatch && r.URL.Path == "/v3/templates/d-test-123":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if tmpl, ok := templates["d-test-123"]; ok {
				if name, exists := body["name"]; exists {
					tmpl["name"] = name
				}
				tmpl["updated_at"] = "2026-01-02 00:00:00"
				templates["d-test-123"] = tmpl
				json.NewEncoder(w).Encode(tmpl)
			}

		case r.Method == http.MethodDelete && r.URL.Path == "/v3/templates/d-test-123":
			delete(templates, "d-test-123")
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
						name = "my-template"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template.test", "id", "d-test-123"),
					resource.TestCheckResourceAttr("sendgrid_template.test", "name", "my-template"),
					resource.TestCheckResourceAttr("sendgrid_template.test", "generation", "dynamic"),
				),
			},
			{
				Config: `
					resource "sendgrid_template" "test" {
						name = "updated-name"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template.test", "name", "updated-name"),
				),
			},
			{
				ResourceName:      "sendgrid_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
