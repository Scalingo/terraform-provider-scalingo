# Initialize the scalingo provider with a token generated for your user
provider "scalingo" {
  api_token = "tk-us-1234567890"
  region    = "osc-fr1"
}

# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name   = "my-awesome-app"
}

# Provision a highly available PostgreSQL cluster and attach it to the application
resource "scalingo_addon" "my-db" {
  provider_id = "postgresql"
  plan = "postgresql-business-1024"
  app = "${scalingo_app.my-app.id}"
}

# Configure domain 'example.com' to be targeting your application
resource "scalingo_domain" "my-domain" {
  common_name = "example.com"
  app = "${scalingo_app.my-app.id}"
}
