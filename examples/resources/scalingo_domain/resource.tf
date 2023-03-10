resource "scalingo_app" "test_app" {
  name = "terraform-testapp"
}

# Create the canonical domain for an app
resource "scalingo_domain" "wwwtestappcom" {
  common_name = "www.testapp.com"
  app         = scalingo_app.test_app.id
  canonical   = true
}

# Create an alias domain for an app
resource "scalingo_domain" "testappcom" {
  common_name = "testapp.com"
  app         = scalingo_app.test_app.id
  canonical   = false
}
