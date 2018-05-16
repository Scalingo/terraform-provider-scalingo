---
layout: "scalingo"
page_title: "Scalingo: scalingo_app"
sidebar_current: "docs-scalingo-resource-domain"
description: |-
  Provides a Scalingo Domain resource. This can be used to add a custom domain name on an applications on Scalingo.
---

# scalingo_app

Provides a Scalingo App resource. This can be used to
create and manage applications on Scalingo.

## Example Usage

```terraform
# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name = "my-awesome-app"
}

# Associate the custom domain
resource "scalingo_domain" "my-domain" {
  common_name = "www.mydomain.com"
  app = "${scalingo_app.my-app.id}"
}
```

## Argument Reference

The following arguments are supported:

* `common_name` - (Required) The domain name to serve requests from.
* `app` - (Required) The Scalingo app ID to link to.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the of the domain record.
- `hostname` - The domain name traffic will be served as.

