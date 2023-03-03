data "scalingo_region" "osc-fr1" {
  name = "osc-fr1"
}

output "scalingo_dashboard" {
  value = scalingo_region.osc-fr1.dashboard
}