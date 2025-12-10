resource "scalingo_database" "test_postgres" {
  name       = "my-postgres-db"
  technology = "postgresql-ng"
  plan       = "postgresql-ng-enterprise-4096"
}

output "app_id" {
  value = scalingo_database.test_postgres.app_id
}

output "database_id" {
  value = scalingo_database.test_postgres.database_id
}
