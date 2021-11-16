package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/Scalingo/terraform-provider-scalingo/scalingo"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scalingo.Provider,
	})
}
