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
  name = "terraform-addon-dbfeature"
}

resource "scalingo_addon" "test_redis" {
  provider_id       = "scalingo-redis"
  plan              = "redis-sandbox"
  app               = scalingo_app.test_app.id
  database_features = ["force-ssl", "redis-aof"]
}
