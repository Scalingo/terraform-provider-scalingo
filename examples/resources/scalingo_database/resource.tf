resource "scalingo_database" "test_postgres" {
  name       = "my-postgres-db"
  technology = "postgresql-ng"
  plan       = "postgresql-ng-enterprise-4096"
}
