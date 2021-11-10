package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"terraform-provider-qovery/qovery"
)

func main() {
	tfsdk.Serve(context.Background(), qovery.New, tfsdk.ServeOpts{
		Name: "qovery",
	})
}
