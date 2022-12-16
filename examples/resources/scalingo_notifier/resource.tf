resource "scalingo_app" "test_app" {
  name = "terraform-test-alert"
}

data_source "scalingo_notification_platform" "slack" {
  name = "slack"
}

# Create a Slack based notifier to get notifications for all events on the
# application
resource "scalingo_notifier" "all_events" {
  app             = scalingo_app.test_app.id
  platform_id     = data_source.scalingo_notification_platform.slack.id
  webhook_id      = "https://hooks.slack.com/services/..."

  name            = "Email Audit Notifier"
  send_all_events = true
  emails          = ["ops@example.com"]
}
