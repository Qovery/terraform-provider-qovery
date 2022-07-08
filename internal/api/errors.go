package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type apiError struct {
	err        error
	action     apiAction
	resource   apiResource
	resourceID string
	resp       *http.Response
}

type apiErrorPayload struct {
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

func (e apiError) IsNotFound() bool {
	// NOTE: consider 403 Forbidden as a 404 NotFound until the api is fixed
	return e.resp.StatusCode == http.StatusNotFound ||
		e.resp.StatusCode == http.StatusForbidden
}

func (e apiError) IsBadRequest() bool {
	// NOTE: consider 403 Forbidden as a 404 NotFound until the api is fixed
	return e.resp.StatusCode == http.StatusBadRequest
}

func (e apiError) Error() string {
	return e.Detail()
}

func (e apiError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

func (e apiError) Detail() string {
	var extra string
	payload := e.errorPayload()

	if e.err != nil {
		extra = fmt.Sprintf("unexpected error: %s", e.err)
		if payload != nil && payload.Message != "" {
			extra = fmt.Sprintf("unexpected error: %s - %s", e.err, payload.Message)
		}
	} else {
		extra = fmt.Sprintf("unexpected status code: %d", e.resp.StatusCode)
	}
	return fmt.Sprintf("Could not %s %s '%s', %s", e.action, e.resource, e.resourceID, extra)
}

func (e apiError) errorPayload() *apiErrorPayload {
	if e.err == nil || e.resp == nil {
		return nil
	}

	body, err := io.ReadAll(e.resp.Body)
	if err != nil {
		return nil
	}

	var payload apiErrorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	return &payload
}

func newApiErrorFromError(err error) *apiError {
	switch err.(type) {
	case *apiError:
		return err.(*apiError)
	default:
		return nil
	}
}

func IsErrNotFound(err error) bool {
	apiErr := newApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsNotFound()
}

func IsErrBadRequest(err error) bool {
	apiErr := newApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsBadRequest()
}

func newError(action apiAction, resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return &apiError{
		err:        err,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		resp:       resp,
	}
}

func newCreateError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionCreate, resource, resourceID, resp, err)
}

func newReadError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionRead, resource, resourceID, resp, err)
}

func newUpdateError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionUpdate, resource, resourceID, resp, err)
}

func newDeleteError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionDelete, resource, resourceID, resp, err)
}

func newStopError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionStop, resource, resourceID, resp, err)
}

func newRestartError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionRestart, resource, resourceID, resp, err)
}

func newDeployError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newError(apiActionDeploy, resource, resourceID, resp, err)
}
