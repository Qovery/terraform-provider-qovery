package qovery

import "github.com/hashicorp/terraform-plugin-framework/types"

type ExecutionCommand struct {
	Entrypoint types.String   `tfsdk:"entrypoint"`
	Arguments  []types.String `tfsdk:"arguments"`
}
