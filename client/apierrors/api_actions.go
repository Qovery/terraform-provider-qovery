package apierrors

type APIAction string

const (
	APIActionCreate       APIAction = "create"
	APIActionRead         APIAction = "read"
	APIActionUpdate       APIAction = "update"
	APIActionDelete       APIAction = "delete"
	APIActionDeploy       APIAction = "deploy"
	APIActionStop         APIAction = "stop"
	APIActionRedeploy     APIAction = "redeploy"
	APIActionNotSupported APIAction = "not supported"
)
