resource "scalingo_app" "test_app" {
  name = "terraform-testapp"

  environment = {
    VARIABLE1 = "Value 1",
    VARIABLE2 = "Value 2"
  }

  force_https = true
}
