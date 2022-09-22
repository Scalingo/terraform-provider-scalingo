variable "api_token" {}

terraform {
  required_providers {
    scalingo = {
      source = "Scalingo/scalingo"
    }
  }
}

provider "scalingo" {
  api_token = var.api_token
}

data "scalingo_invoices" "all" {
  after  = "2021-01-01"
  before = "2022-01-01"
}

output "prices" {
  description = "All Prices"
  value       = data.scalingo_invoices.all.invoices.*.total_price
}

output "dates" {
  description = "All billing months"
  value       = data.scalingo_invoices.all.invoices.*.billing_month
}
