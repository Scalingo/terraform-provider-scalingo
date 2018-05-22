---
layout: "scalingo"
page_title: "Scalingo: scalingo_github_link"
sidebar_current: "docs-scalingo-resource-github-link"
description: |-
  Provides a Scalingo Github resource. These can be attached to a Scalingo app.
---

# scalingo_github_link

Provides a Scalingo Github Link resource. These must be attached to a Scalingo app.

## Example Usage

```terraform
# Create a new Scalingo app
resource "scalingo_app" "my-app" {
  name   = "my-awesome-app"
}

resource "scalingo_github_link" "samplegomartini" {
  app                             = "${scalingo_app.my-app.id}"
  source                          = "https://github.com/johnsudaar/sample-go-martini"
  branch                          = "master"
  auto_deploy                     = true
  review_apps                     = true
  deploy_on_branch_change         = true
  destroy_review_app_on_close     = true
  destroy_stale_review_app        = true
  destroy_closed_review_app_after = 2
  destroy_stale_review_app_after  = 4
}
```

## Argument Reference

The following arguments are supported:

* `app` - (Required) The Scalingo app to add to.
* `source` - (Required) Address of the Github repository
* `branch` - (Optional) Name of the branch used for the `auto_deploy` and `deploy_on_branch_change` features
* `auto_deploy` - (Optional) Start a new deployment when new commits are added to the selected branch
* `review_apps` - (Optional) Enable deployment of [review apps](https://doc.scalingo.com/platform/app/review-apps)
* `destroy_review_app_on_close` - (Optional) Enable review app deletion when the pull request is closed
* `destroy_stale_review_app_after` - (Optional) Time between the pull request close and the review app deletion
* `destroy_stale_review_app` - (Optional) Enable review apps deletion when there is no new commit in the selected pull request
* `destroy_stale_review_app_after` - (Optional) Time between two commits needed to consider this review_app staled
* `deploy_on_branch_change` - (Optional) Send a manual deploy when the branch change (or on initial resource creation)
