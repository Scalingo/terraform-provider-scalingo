resource "scalingo_app" "test_app" {
  name = "terraform-addon"
}

resource "scalingo_addon" "test_redis" {
  provider_id       = "scalingo-redis"
  plan              = "redis-starter-256"
  app               = scalingo_app.test_app.id
  database_features = ["force-ssl", "redis-aof"]
}
