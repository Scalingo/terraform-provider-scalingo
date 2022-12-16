resource "scalingo_app" "test_app" {
  name = "terraform-test-autoscaler"
}

# Create an autoscaler to scale 'web' containers to ensure CPU consumption stays under 80%
resource "scalingo_autoscaler" "test_autoscaler" {
  app            = scalingo_app.test_app.id
  container_type = "web"
  min_containers = 2
  max_containers = 10
  metric         = "cpu"
  target          = 0.8
}
