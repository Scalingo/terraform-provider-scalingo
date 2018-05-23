---
layout: "scalingo"
page_title: "Scalingo: scalingo_collaborator"
sidebar_current: "docs-scalingo-resource-collaborator"
description: |-
  Provides a Scalingo Collaborator resource. These can be attached to a Scalingo app.
---

# scalingo_collaborator

Provides a Scalingo Collaborator resource. These can be attached to a Scalingo app.

## Example Usage

```terraform
# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name   = "my-awesome-app"
}

resource "scalingo_collaborator" "johndoe" {
  app   = "${scalingo_app.my-app.id}"
  email =  "john.doe@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `app` - (Required) The scalingo app to add the collaborator to
* `email` - (Required) Email of the collaborator

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the of the collaborator record.
- `email` - Email of the collaborator
- `username` - Username of the collaborator (or `n/a` if the invitation is pending)
- `status` - `pending` if the invitation not yet accepted or `accepted` if the invitation has been accepted
