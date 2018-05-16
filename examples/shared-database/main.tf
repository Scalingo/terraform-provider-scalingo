resource "scalingo_app" "test_app" {
  name = "terraform-testapp"

  environment {
    TEST_VAR = "test_var"
  }
}

resource "scalingo_app" "test_app_fr" {
  name = "terraform-testapp-fr"

  environment {
    MY_DB = "${lookup(scalingo_app.test_app.all_environment, "SCALINGO_REDIS_URL", "n/c")}"
  }
}

resource "scalingo_domain" "wwwtestappcom" {
  common_name = "www.testapp.com"
  app         = "${scalingo_app.test_app.id}"
}

resource "scalingo_addon" "test_redis" {
  provider_id = "scalingo-redis"
  plan        = "free"
  app         = "${scalingo_app.test_app.id}"
}
