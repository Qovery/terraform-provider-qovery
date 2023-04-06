package qovery

type ExecutionCommand struct {
	Entrypoint *string  `tfsdk:"entrypoint"`
	Arguments  []string `tfsdk:"arguments"`
}
