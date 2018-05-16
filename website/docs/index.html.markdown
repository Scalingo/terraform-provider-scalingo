---
layout: "scalingo"
page_title: "Provider: Scalingo"
sidebar_current: "docs-scalingo-index"
description: |-
  The Template provider is used to template strings for other Terraform resources.
---

# Scalingo Provider

The Scalingo provider is used to interact with the resources supported by Scalingo. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.


## Example Usage

```terraform
# Configure the Scalingo provider
provider "scalingo" {
  api_key = "${var.scalingo_api_key}"
}

# Create a new application
resource "scalingo_app" "my_app" {
  # ...
}
```

## Argument reference

The following arguments are supported:

- `api_key` - (Required) Scalingo API token. This can also be sourced from the `SCALINGO_API_TOKEN` environment variable.
- `api_url` - (Optional) URL of the Scalingo API. This can also be sourced from the `SCALINGO_API_URL` environment variable. If not set, this will default to `https://api.scalingo.com/`.
- `auth_url` - (Optional) URL of Scalingo's authentication API. This can also be sourced from the `SCALINGO_AUTH_URL` environment variable. If not set, this will default to `https://auth.scalingo.com/`


