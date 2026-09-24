variable "database_id" {
  type        = string
  description = "ID of the Scalingo database to read endpoints from"
}

data "scalingo_database_endpoints" "all" {
  database_id                 = var.database_id
  include_default_credentials = true
}

data "scalingo_database_endpoints" "public" {
  database_id                 = var.database_id
  type                        = "public-rw"
  include_default_credentials = true
}

output "database_endpoint_hostnames" {
  value = [for endpoint in data.scalingo_database_endpoints.all.endpoints : endpoint.hostname]
}

output "public_endpoint_hostname" {
  value = data.scalingo_database_endpoints.public.endpoints.0.hostname
}

output "public_endpoint_passwords" {
  value = data.scalingo_database_endpoints.public.endpoints.0.password
  sensitive = true
}
