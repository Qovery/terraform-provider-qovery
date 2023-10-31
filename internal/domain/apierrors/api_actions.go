package apierrors

// APIAction is an enum that contains every type of actions done using the api.
// This is used to generate a detailed error message displayed by terraform when the api return an error.
type APIAction string

const (
	APIActionCreate   APIAction = "create"
	APIActionRead     APIAction = "read"
	APIActionUpdate   APIAction = "update"
	APIActionDelete   APIAction = "delete"
	APIActionDeploy   APIAction = "deploy"
	APIActionStop     APIAction = "stop"
	APIActionRedeploy APIAction = "redeploy"
)
