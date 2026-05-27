# terraform-provider-sendgrid

Terraform provider for managing SendGrid email templates and images as code.

## Resources

| Resource | Description |
|----------|-------------|
| `sendgrid_template` | Manages a SendGrid dynamic template |
| `sendgrid_template_version` | Manages an immutable version of a template |
| `sendgrid_image` | Manages an image on SendGrid's CDN |

## Usage

```hcl
terraform {
  required_providers {
    sendgrid = {
      source = "way-platform/sendgrid"
    }
  }
}

provider "sendgrid" {
  # Set via SENDGRID_API_KEY environment variable
}

resource "sendgrid_image" "logo" {
  file_path   = "${path.module}/assets/logo.png"
  file_sha256 = filesha256("${path.module}/assets/logo.png")
}

resource "sendgrid_template" "welcome" {
  name = "Welcome Email"
}

resource "sendgrid_template_version" "welcome" {
  template_id  = sendgrid_template.welcome.id
  name         = "welcome:${filesha256("${path.module}/templates/welcome.html")}"
  subject      = "Welcome to {{company_name}}"
  html_content = templatefile("${path.module}/templates/welcome.html", {
    logo_url = sendgrid_image.logo.url
  })
}
```

## Development

See [AGENTS.md](AGENTS.md) for build instructions and conventions.

```bash
mise install
mise run build
```

## License

MIT - Way Data Technologies Oy
