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
  name            = "Slack Audit Notifier"
  platform_id     = data_source.scalingo_notification_platform.slack.id
  
  active          = true
  send_all_events = true
  webhook_url     = "https://hooks.slack.com/services/..."
}

# Create a notifier to get emails for all alerts and only selected critical
# events on the application
data_source "scalingo_notification_platform" "email" {
  name = "email"
}

resource "scalingo_notifier" "all_events" {
  app             = scalingo_app.test_app.id
  name            = "Email Audit Notifier"
  platform_id     = data_source.scalingo_notification_platform.email.id
  
  send_all_alerts = true
  selected_events = [
    "addon_deleted",
    "addon_suspended",
    "app_crashed_repeated",
    "domain_removed",
    "notifier_removed",
    "variable_removed",
  ]
  emails          = ["ops@example.com"]
}
