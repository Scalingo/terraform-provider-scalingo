resource "scalingo_app" "test_app" {
  name = "terraform-test-firewall_rules"
}

resource "scalingo_app_firewall_rule" "office" {
  app   = scalingo_app.test_app.id
  cidr  = "192.0.2.0/24"
  label = "office"
}

resource "scalingo_app_firewall_rule" "single_ip" {
  app  = scalingo_app.test_app.id
  cidr = "198.51.100.4/32"
}
