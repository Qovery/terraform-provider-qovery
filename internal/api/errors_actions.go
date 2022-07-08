package api

type apiAction string

const (
	apiActionCreate  apiAction = "create"
	apiActionRead    apiAction = "read"
	apiActionUpdate  apiAction = "update"
	apiActionDelete  apiAction = "delete"
	apiActionDeploy  apiAction = "deploy"
	apiActionStop    apiAction = "stop"
	apiActionRestart apiAction = "restart"
)
