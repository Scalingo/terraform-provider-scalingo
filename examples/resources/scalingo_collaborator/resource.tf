resource "scalingo_app" "test_app" {
  name = "terraform-testapp"
}

locals {
  team = ["dev@example.com", "ops@example.com"]
}

resource "scalingo_collaborator" "collaborators" {
  for_each = toset(local.team)

  app      = scalingo_app.test_app.id
  email    = each.key
}
