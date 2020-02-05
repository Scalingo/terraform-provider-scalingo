Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Maintainers
-----------

This provider plugin is maintained by the [Scalingo](https://scalingo.com) team.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.9 (to build the provider plugin)

Usage
---------------------

### Configuring with variables

```
variable "scalingo_api_token" {}

provider "scalingo" {
  api_token = "${var.scalingo_api_token}"
  region = "osc-fr1"
}
```

```bash
export TF_VAR_scaling_api_token=tk-us-1234567890
terraform plan
```

### Configuration with environment variables

```
provider "scalingo" {}
```

```bash
export SCALINGO_REGION=osc-fr1
export SCALINGO_API_TOKEN=tk-us-1234567890

terraform plan
```



Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/Scalingo/terraform-provider-scalingo`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/Scalingo
$ git clone git@github.com:Scalingo/terraform-provider-scalingo
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/Scalingo/terraform-provider-scalingo
$ make build
```

Using the provider
----------------------
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (version 1.9+ is
*required*). You'll also need to correctly setup a
[GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding
`$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put
the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-scalingo
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

### Testing the plugin against the staging environment

Please read the README in the `scalingo-terraform` to get more information about
that.
