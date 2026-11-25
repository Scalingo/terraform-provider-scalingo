resource "scalingo_project" "test_project" {
  name    = "test-project"
  default = true
}

# The organization slug is used to create the project in an organization.
resource "scalingo_project" "organization_project" {
  name         = "organization-project"
  organization = "my-organization"
}
