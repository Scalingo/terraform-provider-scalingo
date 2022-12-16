resource "scalingo_app" "test_app" {
  name = "terraform-testapp"
}

resource "scalingo_domain" "wwwtestappcom" {
  common_name = "www.testapp.com"
  app         = scalingo_app.test_app.id
  canonical   = false
}
