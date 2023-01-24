---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "scalingo_notification_platform Data Source - terraform-provider-scalingo"
subcategory: ""
description: |-
  Notification platforms are the different destination to which notifications and alerts about an application can be sent
---

# scalingo_notification_platform (Data Source)

Notification platforms are the different destination to which notifications and alerts about an application can be sent

## Example Usage

```terraform
# Get "email" notification platform to create notifiers based on it
data_source "scalingo_notification_platform" "email" {
  name = "email"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Slug name of the notification platform

### Optional

- `available_event_ids` (Set of String) List of event IDs which can be sent through this platform

### Read-Only

- `description` (String) Textual description
- `display_name` (String) Human-enriched name of the notification platform
- `id` (String) The ID of this resource.
- `logo_url` (String) Logo image URL representing the notification platform

