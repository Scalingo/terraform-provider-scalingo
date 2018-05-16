---
layout: "scalingo"
page_title: "Scalingo: scalingo_app"
sidebar_current: "docs-scalingo-resource-app"
description: |-
  Provides a Scalingo App resource. This can be used to create and manage applications on Scalingo.
---

# scalingo_app

Provides a Scalingo App resource. This can be used to
create and manage applications on Scalingo.

## Example Usage

```terraform
# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name = "my-awesome-app"

  environment {
    FOO = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application. In Heroku, this is also the
   unique ID, so it must be unique and have a minimum of 3 characters.
* `environment` - (Optional) Configuration variables for the application.
     The config variables in this map are not the final set of configuration
     variables, but rather variables you want present. That is, other
     configuration variables set externally won't be removed by Terraform
     if they aren't present in this list.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the app. This is also the name of the application.
* `name` - The name of the application.
* `git_url` - The Git URL for the application. This is used for
   deploying new versions of the app.
* `url` - The (HTTP) URL that the application can be accessed
   at by default.
* `all_environment` - A map of all of the configuration variables that
    exist for the app, containing both those set by Terraform and those
    set externally.

