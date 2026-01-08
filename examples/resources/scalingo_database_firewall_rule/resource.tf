# Custom range firewall rule
resource "scalingo_database_firewall_rule" "office" {
  database_id = scalingo_database.my_db.database_id
  cidr        = "203.0.113.0/24"
  label       = "Office network"
}

# Managed range firewall rule (e.g., allow traffic from Scalingo region)
data "scalingo_database_firewall_managed_range" "region" {
  database_id = scalingo_database.my_db.database_id
  name        = "Scalingo osc-fr1 region"
}

resource "scalingo_database_firewall_rule" "scalingo_region" {
  database_id      = scalingo_database.my_db.database_id
  managed_range_id = data.scalingo_database_firewall_managed_range.region.id
}
