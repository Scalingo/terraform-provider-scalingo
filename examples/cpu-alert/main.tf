variable "api_token" {}

terraform {
  required_providers {
    scalingo = {
      source = "Scalingo/scalingo"
    }
  }
}

provider "scalingo" {
  api_token = var.api_token
}

resource "scalingo_app" "test_app" {
  name = "terraform-test-alert"
}

resource "scalingo_alert" "test_alert" {
  app            = scalingo_app.test_app.id
  container_type = "web"
  metric         = "cpu"
  limit          = 0.8
}

