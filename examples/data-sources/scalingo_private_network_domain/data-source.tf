data "scalingo_private_network_domain" "example" {
  app       = "app-12345678-1234-1234-1234-1234567890ab"
  page      = 1
  page_size = 50
}
