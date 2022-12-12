resource "scalingo_app" "test_app" {
  name = "terraform-testapp"
}

resource "scalingo_collaborator" "customer" {
  app   = scalingo_app.test_app.id
  email = "customer@scalingo.com"
}
