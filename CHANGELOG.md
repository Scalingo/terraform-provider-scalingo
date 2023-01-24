# Changelog

## To be Released

## v1.0.3

* resource(addon): Property `database_features` is now considered as computed
  since backend is injecting default features, users should not have to list
  all of them. #119

## v1.0.2

* resource(app): Fix change of `stack_id` which was not applied correctly #117

## v1.0.1

* Add `Description` for all Resources and Data Sources
* Create various examples in `examples` directory to generate appropriate documentation on Terraform registry website:
  https://registry.terraform.io/providers/Scalingo/scalingo/latest/docs

## v1.0.0

### Breaking Changes

* Resource `github_link`: `deploy_on_branch_change` has been removed
* Resource `run`: complete resource has been removed

### Deprecation

* Resource `github_link` has been deprecated in favor of `scm_repo_link`

### Changes

* resource(addon): Add `database_features` property to enable automatically database addon features #95
* resource(domain): Add `canonical` property to configure a Domain as canonical #80
* resource(alert): Add `alert` resource to configure application alerts on metrics #86 #93
* resource(scm_repo_link): Add `scm_repo_link` resource to create SCM links with application and configure it #85 #65
* resource(scm_integration): Add `scm_integration` to create GitHub Enterprise and Gitlab self-hosted SCM Integration #46
* resource(ssh_key): Add `ssh_key` resource to handle use public SSH keys #72
* resource(run): Drop `run` resource: not pertinent in terraform provider #70
* resource(github_link): Deprecate `github_link` resource entirely
* resource(github_link): Remove `deploy_on_branch_change` property to prevent resource creation side effect #77
* data_provider(addon_provider): Add `addon_provider` data provider to list accessible addon provider #68
* data_provider(container_size): Add `container_size` data provider to get metadata about every container size #76
* data_provider(region): Add data `region` provider to get metadata for every accessible region #69
* data_provider(scm_integration): Add `scm_integrations` data provider to get information from GitHub and Gitlab #85 #66
* data_provider(invoices): Add `invoices` data provider to get invoices from an account #87
* deps(go-scalingo): Bump v6.0.0 #95
* deps(terraform-plugin-sdk): Bump v2.23 #96
* chore(refactoring) Use generics for filters #78

## v0.5.2

* fix(app) Do not try to remove the stack_id if it's not set in the app resource

## v0.5.1

* fix(notifiers) Rename webhook url field to webhook_url

## v0.5.0

* feature(log-drains): Add support for log drains
* chore(errors): Switch error management to fmt.Errorf
* feature(stacks): Add support for Scalingo stacks
* chore(go-scalingo): Upgrade to go-scalingo v4
* chore(terraform-plugin-sdk): Upgrade to SDK v2

## v0.4.2

* chore(go): use go 1.17
* feat(force-https): add support for force-https setting
* Bump github.com/hashicorp/terraform-plugin-sdk from 1.17.0 to 1.17.2

## 0.4.1

* Update terraform SDK from 1.7.0 to 1.17.0

## 0.4.0

* Migration to official terraform SDK
  https://www.terraform.io/docs/extend/guides/v1-upgrade-guide.html

## 0.3.0

* Compatibility with Terraform 0.13

## 0.2.0

* Compatibility with terraform 0.12.20

## 0.1.0

* Project is used in production by the Scalingo core team for 18 months
* Update of `github.com/Scalingo/go-scalingo` to take into account regions

## 0.0.1

* Initial release
