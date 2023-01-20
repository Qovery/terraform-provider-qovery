package apierrors

import (
	"net/http"
)

func NewError(action APIAction, resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return &APIError{
		err:        err,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		res:        res,
	}
}

func NewCreateError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionCreate, resource, resourceID, res, err)
}

func NewReadError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionRead, resource, resourceID, res, err)
}

func NewUpdateError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionUpdate, resource, resourceID, res, err)
}

func NewDeleteError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionDelete, resource, resourceID, res, err)
}

func NewStopError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionStop, resource, resourceID, res, err)
}

func NewRedeployError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionRedeploy, resource, resourceID, res, err)
}

func NewDeployError(resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	return NewError(APIActionDeploy, resource, resourceID, res, err)
}
