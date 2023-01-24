---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "scalingo_app Resource - terraform-provider-scalingo"
subcategory: ""
description: |-
  Resource representing an application
---

# scalingo_app (Resource)

Resource representing an application

## Example Usage

```terraform
resource "scalingo_app" "test_app" {
  name = "terraform-testapp"

  environment = {
    VARIABLE1 = "Value 1",
    VARIABLE2 = "Value 2"
  }

  force_https = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application

### Optional

- `environment` (Map of String) Key-value map of environment variables attached to the application
- `force_https` (Boolean) Redirect HTTP traffic to HTTPS + HSTS header if enabled
- `stack_id` (String) ID of the base stack to use (scalingo-18/scalingo-20)

### Read-Only

- `all_environment` (Map of String) Computed key-value map containing environment in read-only
- `git_url` (String) Hostname to use to deploy code with Git + SSH
- `id` (String) The ID of this resource.
- `url` (String) Base URL (https://*) to access the application

