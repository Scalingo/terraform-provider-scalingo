# Terraform Provider

[Documentation](https://registry.terraform.io/providers/Scalingo/scalingo/latest/docs)

## Maintainers

This provider plugin is maintained by the [Scalingo](https://scalingo.com) team.

## Requirements

-	[Terraform](https://developer.hashicorp.com/terraform/downloads)
-	[Go](https://go.dev/doc/install) to build the provider plugin

## Usage

### Configuring with variables

```
variable "scalingo_api_token" {}

provider "scalingo" {
  api_token = "${var.scalingo_api_token}"
  region = "osc-fr1"
}
```

```bash
export TF_VAR_scalingo_api_token=tk-us-1234567890
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

### Use terraform import

Some resources are using a specific syntax using a `:` to provide multiple information.

The terraform import command is formatted like the following:

```bash
terraform import <ADDR> <ID>
```

For example, to import alerts the ID is composed like: `<application ID>:<alert ID>`

```bash
terraform import scalingo_alert.cpu_alert my-app:al-18f30d13-3c19-422d-a0d6-6cdb254baeb7
```


## Building The Provider

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

## Using the provider

See [documentation](https://registry.terraform.io/providers/Scalingo/scalingo/latest/docs)

## Developing the Provider

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (version 1.9+ is
*required*). You'll also need to correctly setup a
[GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding
`$GOPATH/bin` to your `$PATH`.

### Building

To compile the provider, run `make build`. This will build the provider and put
the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-scalingo
...
```

### Testing

In order to run the tests of the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

### Testing the plugin against the development environment

First you will need to create a terraform configuration file defining the plugin installation path

```
# local_dev.tfrc

provider_installation {
  dev_overrides {
    "scalingo/scalingo" = "<your-gopath>/bin/"
  }
}
```

Then you can export the configuration path in the `TF_CLI_CONFIG_FILE` environment variable.

```sh
export TF_CLI_CONFIG_FILE=./local_dev.tfrc
```

Alternatively you can add the provider configuration to the `$HOME/.terraformrc` file.

Now you need to ensure to have run `make build` to have the local changes built and ready for testing.
When this is done, you can start writing your own Terraform file. For example:

```
terraform{
    required_providers{
        scalingo={
            source="scalingo/scalingo"
            version="2.6.0"
        }
    }
}

provider "scalingo"{
    region = ""
    auth_api_url = <INSERT LOCAL AUTH URL>
    api_url = <INSERT LOCAL API URL>
    db_api_url = <INSERT LOCAL DB API URL>
    api_token = <INSERT TOKEN>
}

resource "scalingo_app" "test-app" {
    name = "test-application-terraform"
}
```

Then you just need to run the terraform commands:
```sh
terraform plan
terraform apply
...
```

### Testing the plugin against the staging environment

In order to target staging, there is only one change to perform. The scalingo provider
from your terraform file should look like:

```
provider "scalingo"{
    region = ""
    auth_api_url = "https://auth.st-sc.fr"
    api_url = "https://api.osc-st-fr1.st-sc.fr"
    db_api_url = "https://db-api.osc-st-fr1.st-sc.fr"
    api_token = "<INSERT TOKEN>"
}

```

### Generate Documentation

Documentation of the provider is based on official Terraform documentation
plugin: https://github.com/hashicorp/terraform-plugin-docs

```
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs
```

By running the process, it will scan all resources metadata plus the examples
directory to generate a complete documentation structure in the `docs/`
directory.

## Release a New Version

Instructions on this [Notion page](https://www.notion.so/scalingooriginal/New-Terraform-Provider-Release-40cd0af66b1f48148fb641ea138a22e5).
