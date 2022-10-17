package apierrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// apiErrorPayload represents the error payload that comes from Qovery's API client.
// It is used to get the actual error message from the api.
type apiErrorPayload struct {
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

// ApiError represents an error that comes from Qovery's API client.
// It contains all the information needed to understand an api error.
type ApiError struct {
	err        error          // err is the actual error returned by the api client.
	action     ApiAction      // action that produced the api client error.
	resource   ApiResource    // resource that produced the api client error.
	resourceID string         // resourceID that produced the api client error. [NOTE: it is replaced by the resource name in some cases (for failed create requests)]
	resp       *http.Response // resp is the response returned by the api client.
}

// IsNotFound returns weather the error is a 404 or not.
// NOTE: Since the api returns a 403 when a resource is not found, we also consider those as a 404.
func (e ApiError) IsNotFound() bool {
	if e.resp == nil {
		return false
	}

	// NOTE: consider 403 Forbidden as a 404 NotFound until the api is fixed
	return e.resp.StatusCode == http.StatusNotFound ||
		e.resp.StatusCode == http.StatusForbidden
}

// IsBadRequest returns weather the error is a 400 or not.
func (e ApiError) IsBadRequest() bool {
	if e.resp == nil {
		return false
	}

	return e.resp.StatusCode == http.StatusBadRequest
}

// Error implements the Error interface.
// It returns the detailed error message for this ApiError.
func (e ApiError) Error() string {
	return e.Detail()
}

// Summary returns a brief description of the error with the ApiResource and ApiAction.
func (e ApiError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

// Detail return a detailed error message for this ApiError.
// It tries to read the error payload received from the api client to give extra information about the error.
func (e ApiError) Detail() string {
	var extra string
	payload := e.errorPayload()

	if e.err != nil {
		extra = fmt.Sprintf("unexpected error: %s", e.err)
		if payload != nil && payload.Message != "" {
			extra = fmt.Sprintf("%s - %s", extra, payload.Message)
		}
	} else {
		extra = fmt.Sprintf("unexpected status code: %d", e.resp.StatusCode)
	}

	return fmt.Sprintf("Could not %s %s '%s', %s", e.action, e.resource, e.resourceID, extra)
}

// errorPayload tries to read the response body to extract the error payload sent by the api client.
// It returns nil if the body is empty.
func (e ApiError) errorPayload() *apiErrorPayload {
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

// NewApiErrorFromError tries to cast an error into an ApiError.
// This is useful when working with ApiError passed as an `error` type to get the actual ApiError type.
func NewApiErrorFromError(err error) *ApiError {
	switch err.(type) {
	case *ApiError:
		return err.(*ApiError)
	default:
		return nil
	}
}

// IsErrNotFound takes an error type and tries to cast it into a ApiError to check weather the error is an 404 or not.
// It returns false if the casting fails.
func IsErrNotFound(err error) bool {
	apiErr := NewApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsNotFound()
}

// IsErrBadRequest takes an error type and tries to cast it into a ApiError to check weather the error is an 400 or not.
// It returns false if the casting fails.
func IsErrBadRequest(err error) bool {
	apiErr := NewApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsBadRequest()
}

// NewApiError returns a new instance of ApiError with the given parameters.
func NewApiError(action ApiAction, resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return &ApiError{
		err:        err,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		resp:       resp,
	}
}

// NewCreateApiError returns a new instance of ApiError for a `create` action with the given parameters.
func NewCreateApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionCreate, resource, resourceID, resp, err)
}

// NewReadApiError returns a new instance of ApiError for a `read` action with the given parameters.
func NewReadApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionRead, resource, resourceID, resp, err)
}

// NewUpdateApiError returns a new instance of ApiError for an `update` action with the given parameters.
func NewUpdateApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionUpdate, resource, resourceID, resp, err)
}

// NewDeleteApiError returns a new instance of ApiError for a `delete` action with the given parameters.
func NewDeleteApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionDelete, resource, resourceID, resp, err)
}

// NewStopApiError returns a new instance of ApiError for a `stop` action with the given parameters.
func NewStopApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionStop, resource, resourceID, resp, err)
}

// NewRestartApiError returns a new instance of ApiError for a `restart` action with the given parameters.
func NewRestartApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionRestart, resource, resourceID, resp, err)
}

// NewDeployApiError returns a new instance of ApiError for a `deploy` action with the given parameters.
func NewDeployApiError(resource ApiResource, resourceID string, resp *http.Response, err error) *ApiError {
	return NewApiError(ApiActionDeploy, resource, resourceID, resp, err)
}

// NewNotFoundApiError returns a new instance of ApiError for a `not_found` resource with the given parameters.
func NewNotFoundApiError(resource ApiResource, resourceID string) *ApiError {
	return NewApiError(ApiActionRead, resource, resourceID, &http.Response{
		StatusCode: http.StatusNotFound,
	}, errors.New("resource not found"))
}
