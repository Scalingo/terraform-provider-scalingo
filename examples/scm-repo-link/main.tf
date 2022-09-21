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

data "scalingo_scm_integration" "github" {
  scm_type = "github"
}

resource "scalingo_app" "test_app" {
  name = "terraform-test-scm"
}

resource "scalingo_scm_repo_link" "github-link" {
  auth_integration_uuid = data.scalingo_scm_integration.github.id
  app                   = scalingo_app.test_app.id
  source                = "https://github.com/nalpskc/sample-node-express"
}
