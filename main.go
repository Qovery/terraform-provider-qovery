package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/qovery/terraform-provider-qovery/qovery"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have Terraform installed, you can remove the formatting command, but it's suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the documentation generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var version = "dev"

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/Qovery/qovery",
		Debug:   debugMode,
	}

	if err := providerserver.Serve(context.Background(), qovery.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
