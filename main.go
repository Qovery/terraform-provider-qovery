package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"terraform-provider-qovery/qovery"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var version = "dev"

func main() {
	opts := tfsdk.ServeOpts{
		Name: "qovery",
	}

	if err := tfsdk.Serve(context.Background(), qovery.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
