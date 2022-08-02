package apierrors

// ApiAction is an enum that contains every type of actions done using the api.
// This is used to generate a detailed error message displayed by terraform when the api return an error.
type ApiAction string

const (
	ApiActionCreate  ApiAction = "create"
	ApiActionRead    ApiAction = "read"
	ApiActionUpdate  ApiAction = "update"
	ApiActionDelete  ApiAction = "delete"
	ApiActionDeploy  ApiAction = "deploy"
	ApiActionStop    ApiAction = "stop"
	ApiActionRestart ApiAction = "restart"
)
