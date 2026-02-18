package apierrors

import (
	"fmt"
	"io"
	"net/http"
	"time"
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

// NewUnexpectedStateError creates an error for when a resource reaches an unexpected state
func NewUnexpectedStateError(resource APIResource, resourceID string, expected, actual any) *APIError {
	err := fmt.Errorf("%s '%s' reached state %v but expected %v", resource, resourceID, actual, expected)
	return &APIError{
		err:        err,
		action:     APIActionRead,
		resource:   resource,
		resourceID: resourceID,
		res:        nil,
	}
}

// NewUnexpectedClusterStateError creates an error for when a cluster reaches an unexpected state
func NewUnexpectedClusterStateError(orgID, clusterID string, expected, actual any) *APIError {
	err := fmt.Errorf("cluster '%s' in organization '%s' reached state %v but expected %v", clusterID, orgID, actual, expected)
	return &APIError{
		err:        err,
		action:     APIActionRead,
		resource:   APIResourceCluster,
		resourceID: clusterID,
		res:        nil,
	}
}

// NewTimeoutError creates an error for when an operation times out
func NewTimeoutError(timeout time.Duration) *APIError {
	err := fmt.Errorf("operation did not complete within %s", timeout)
	return &APIError{
		err:      err,
		action:   APIActionRead,
		resource: "",
		res:      nil,
	}
}
