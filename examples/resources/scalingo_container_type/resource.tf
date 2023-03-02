resource "scalingo_app" "test_app" {
  name = "terraform-testapp"
}

resource "scalingo_container_type" "web" {
  app    = scalingo_app.test_app.name
  name   = "web"
  amount = 2
  size   = "M"
}