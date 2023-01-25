package main

import (
	"github.com/cv65kr/terraform-provider-temporal/temporal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate tfplugindocs

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: temporal.Provider,
	})
}
