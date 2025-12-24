data "scalingo_database_firewall_managed_range" "osc_fr1" {
  database_id = scalingo_database.my_db.database_id
  name        = "Scalingo osc-fr1 region"
}

output "managed_range_id" {
  value = data.scalingo_database_firewall_managed_range.osc_fr1.id
}
