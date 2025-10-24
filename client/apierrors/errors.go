package apierrors

import (
	"io"
	"net/http"
)

func NewError(action APIAction, resource APIResource, resourceID string, res *http.Response, err error) *APIError {
	var bufferedBody []byte

	// Buffer the response body to prevent EOF errors on subsequent reads
	if res != nil && res.Body != nil {
		body, readErr := io.ReadAll(res.Body)
		if readErr == nil {
			bufferedBody = body
		}
		// Close the original body since we've read it
		res.Body.Close()
	}

	return &APIError{
		err:          err,
		action:       action,
		resource:     resource,
		resourceID:   resourceID,
		res:          res,
		bufferedBody: bufferedBody,
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
