package main

import (
	"github.com/Scalingo/terraform-provider-scalingo/scalingo"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scalingo.Provider,
	})
}
