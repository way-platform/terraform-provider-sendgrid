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
