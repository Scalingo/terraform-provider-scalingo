# Changelog

## To be Released

## v2.3.0

* resource(scalingo_app): expose application's base_url #175
* resource(scalingo_container_type): make amount mandatory #191
* resource(scalingo_domain): Add `letsencrypt_enabled` parameter #205
* resource(scalingo_log_drain): Add `type` parameter mandatory #178
* config: use `SCALINGO_REGION` if environment is set correctly
* chore(deps) Various dependency updates

## v2.2.0

* feat(scalingo_app): add support for router logs and sticky sessions #167
* fix(resource/scalingo_scm_repo_link): the auth_integration_uuid cannot be updated #161
* chore(deps): bump github.com/go-test/deep from v1.0.5 to v1.1.0
* chore(deps): bump github.com/golang/protobuf from 1.5.2 to v1.5.3
* chore(deps): bump github.com/zclconf/go-cty from 1.13.0 to v1.13.1
* chore(deps): bump golang.org/x/net from v0.8.0 to v0.9.0
* chore(deps): bump golang.org/x/sys from v0.6.0 to v0.7.0
* chore(deps): bump golang.org/x/text from v0.8.0 to v0.9.0
* chore(deps): bump google.golang.org/genproto from
  v0.0.0-20230110181048-76db0878b65f to v0.0.0-20230410155749-daa745c078e1
* chore(deps): bump google.golang.org/grpc from 1.53.0 to v1.56.1

## v2.1.0

* resource(collaborator): Re-create collaborator if it has been deleted outside Terraform #137
* resource(notifier): Set default to false for 'send_all_alerts' #139
* resource(scm_repo_link): Add option automatic_creation_from_forks_allowed to enable explicitely creation of review apps from fork #130
* doc(notifier/region): Improve documentation examples #140
* chore(deps): bump github.com/hashicorp/terraform-svchost from 0.0.0-20200729002733-f050f53b9734 to 0.1.0 #135
* chore(deps): bump github.com/golang-jwt/jwt/v4 from 4.4.3 to 4.5.0 #132
* chore(deps): bump google.golang.org/grpc from 1.52.3 to 1.53.0 #136
* chore(deps): bump github.com/hashicorp/yamux #134
* chore(deps): bump github.com/zclconf/go-cty from 1.12.1 to 1.13.0 #133

## v2.0.0

* Remove deprecated `github_link` resource. Use `scm_repo_link` instead
* Update dependencies:
  - github.com/Scalingo/go-scalingo/v6 v6.3.0
  - github.com/hashicorp/terraform-plugin-log v0.7.0
  - github.com/fatih/color v1.14.1
  - github.com/hashicorp/go-plugin v1.4.8
  - github.com/hashicorp/hcl/v2 v2.16.0
  - github.com/hashicorp/terraform-plugin-go v0.14.3
  - github.com/oklog/run v1.1.0
  - golang.org/x/net v0.7.0
  - golang.org/x/sys v0.5.0
  - golang.org/x/text v0.7.0
  - google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6
  - google.golang.org/grpc v1.52.3

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
