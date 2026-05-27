# Agent Instructions

## Overview

Terraform provider for managing SendGrid email templates and images as
infrastructure. Built with the **Terraform Plugin Framework** (not SDKv2).
Minimal scope: 3 resources (`sendgrid_template`, `sendgrid_template_version`,
`sendgrid_image`).

## Commands

Use **mise** for build tasks (run from repo root):

| Command            | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `mise run build`   | Full CI pipeline (download, generate, lint, test, tidy, diff) |
| `mise run lint`    | Run golangci-lint                                             |
| `mise run test`    | Run unit tests                                                |
| `mise run testacc` | Run acceptance tests against mock servers                     |
| `mise run docs`    | Generate Terraform registry documentation                     |

## Architecture

### Code Generation

Go types are generated from the official SendGrid OpenAPI spec
(`twilio/sendgrid-oai`). The pipeline lives in `api/templates/`:

```
01-original.json  (vendored from sendgrid-oai)
     ↓  openapi-overlay (strip paths, keep schemas only)
02-overlayed.json
     ↓  openapi-down-convert (3.1 → 3.0)
03-downconverted.json
     ↓  oapi-codegen (generate Go types)
templates.gen.go
```

Run `mise run generate` to regenerate. Intermediate files (`02-*`, `03-*`)
are gitignored — only the vendored spec and final output are committed.

Use generated types directly everywhere (no aliases). The generated request
body types (e.g. `templates.CreateTemplateJSONRequestBody`) replace hand-written
request structs.

The `/v3/images` endpoints have no published spec — the `Image` struct in
`internal/client/` stays hand-written.

### Client

Thin HTTP client in `internal/client/` using raw `net/http`. One operation per
file (e.g. `create_template.go`, `get_template.go`). Each function takes a
request type (generated or struct) and returns the response type. No SDK
dependencies.

The `Ptr[T]` helper in `ptr.go` is the only generic utility — use it for
optional pointer fields on generated types.

### Resources

Each resource lives in its own package under `internal/resources/`. The resource
file (`resource.go`) and its acceptance test (`resource_test.go`) are colocated.

### Provider

`internal/provider/` handles configuration (API key) and resource registration.
It supports a `SENDGRID_API_URL` env var override to point at mock servers
during tests.

## Key Conventions

- **Go 1.26 idioms**: `new("x")`, `for i := range n`, `min`/`max` builtins
- **Max line length**: 120 characters
- **Testing**: stdlib `testing` + `github.com/google/go-cmp/cmp` only. No testify.
- **Linting**: golangci-lint v2 (see `.golangci.yml`)
- **Build**: mise (see `mise.toml`)
- **Receivers**: 1-2 letter abbreviations (`r` for resource, `c` for client)
- **Errors**: lowercase, no punctuation
- **One operation per file** in client code — request types colocated with their function
- **No type aliases** — use generated types directly, even if verbose

## Testing

### Client Unit Tests

Each client operation has a test file (e.g. `create_template_test.go`) using
`httptest.NewServer` to mock the SendGrid API and verify request/response
handling.

### Acceptance Tests (Resources)

Use `terraform-plugin-testing` with `resource.Test()`. Each resource package
defines its own `testAccProtoV6ProviderFactories` (one-liner) and runs a full
Terraform lifecycle against a local `httptest.NewServer` mock — no real API
credentials needed.

Pattern:
1. Start an `httptest.NewServer` that simulates the relevant SendGrid endpoints
2. Set `SENDGRID_API_KEY=test-key` and `SENDGRID_API_URL=srv.URL` via `t.Setenv`
3. Define `resource.TestStep` entries for create, update (if applicable), and import
4. Assert state with `resource.TestCheckResourceAttr`

Run with `mise run test` (included in `mise run build`) or directly:
```
go test ./internal/resources/...
```

## CI/CD

- **PRs**: GitHub Actions runs `mise run build`
- **Merge to main**: `go-semantic-release` determines version, GoReleaser builds
  multi-platform binaries, publishes GitHub Release, Terraform Registry auto-publishes

## SendGrid API Reference

Templates (documented):
- `POST /v3/templates` — Create
- `GET /v3/templates/{id}` — Read
- `PATCH /v3/templates/{id}` — Update name
- `DELETE /v3/templates/{id}` — Delete

Template Versions (documented):
- `POST /v3/templates/{id}/versions` — Create
- `GET /v3/templates/{id}/versions/{version_id}` — Read
- `DELETE /v3/templates/{id}/versions/{version_id}` — Delete

Images (undocumented, stable 2+ years):
- `POST /v3/images` — Upload (multipart/form-data)
- `GET /v3/images/{id}` — Read
- `DELETE /v3/images/{id}` — Delete
