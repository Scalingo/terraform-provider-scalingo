resource "scalingo_app" "test_app" {
  name = "terraform-test-alert"
}

data_source "scalingo_notification_platform" "email" {
  name = "email"
}

# Create an email based notifier to get alert notifications
resource "scalingo_notifier" "email_alert" {
  app         = scalingo_app.test_app.id
  platform_id = data_source.scalingo_notification_platform.email.id
  name        = "CPU Alert Email Notifications"
  emails      = ["ops@example.com"]
}

# Create an alert which will be triggered if CPU usage is above 80%
resource "scalingo_alert" "test_alert" {
  app            = scalingo_app.test_app.id
  container_type = "web"
  metric         = "cpu"
  limit          = 0.8
  notifiers      = [resource.scalingo_notifier.email_alert.id]
}
