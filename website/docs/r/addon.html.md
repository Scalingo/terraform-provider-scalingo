---
layout: "scalingo"
page_title: "Scalingo: scalingo_addon"
sidebar_current: "docs-scalingo-resource-addon"
description: |-
  Provides a Scalingo Add-On resource. These can be attached to a Scalingo app.
---

# scalingo_addon

Provides a Scalingo Add-On resource. These can be attached to a Scalingo app.

## Example Usage

```terraform
# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name   = "my-awesome-app"
}

resource "scalingo_app" "my-db" {
  provider_id = "scalingo-mongodb"
  plan = "mongo-sandbox"
  app = "${scalingo_app.my-app.id}"
}
```

## Argument Reference

The following arguments are supported:

* `app` - (Required) The Scalingo app to add to.
* `provider_id` - (Required) The provider of the addon to add.
* `plan` - (Required) The plan of the addon to add.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the add-on
* `plan` - The plan name
* `provider_id` - The ID of the plan provider
* `plan_id` - The ID of the plan provided
* `resource_id` - Resource ID of the addon

