package apierrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// apiErrorPayload represents the error payload that comes from Qovery's API client.
// It is used to get the actual error message from the api.
type apiErrorPayload struct {
	Status  int    `json:"status"`
	Message string `json:"detail"`
}

// APIError represents an error that comes from Qovery's API client.
// It contains all the information needed to understand an api error.
type APIError struct {
	err          error          // err is the actual error returned by the api client.
	action       APIAction      // action that produced the api client error.
	resource     APIResource    // resource that produced the api client error.
	resourceID   string         // resourceID that produced the api client error. [NOTE: it is replaced by the resource name in some cases (for failed create requests)]
	Resp         *http.Response // Resp is the response returned by the api client.
	bufferedBody []byte         // bufferedBody is the response body, read once at construction so the error can be inspected multiple times.
}

// IsNotFound returns weather the error is a 404 or not.
// NOTE: Since the api returns a 403 when a resource is not found, we also consider those as a 404.
func (e APIError) IsNotFound() bool {
	if e.Resp == nil {
		return false
	}

	// NOTE: consider 400 Bad Request, 403 Forbidden as a 404 NotFound until the api is fixed
	return (e.Resp.StatusCode == http.StatusBadRequest && strings.Contains(e.Detail(), "exist")) ||
		e.Resp.StatusCode == http.StatusNotFound ||
		e.Resp.StatusCode == http.StatusForbidden
}

// IsBadRequest returns weather the error is a 400 or not.
func (e APIError) IsBadRequest() bool {
	if e.Resp == nil {
		return false
	}

	return e.Resp.StatusCode == http.StatusBadRequest
}

// Error implements the Error interface.
// It returns the detailed error message for this APIError.
func (e APIError) Error() string {
	return e.Detail()
}

// Summary returns a brief description of the error with the APIResource and APIAction.
func (e APIError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

// Detail return a detailed error message for this APIError.
// It tries to read the error payload received from the api client to give extra information about the error.
func (e APIError) Detail() string {
	var extra string
	payload := e.errorPayload()

	if e.err != nil {
		extra = fmt.Sprintf("unexpected error: %s", e.err)
		if payload != nil && payload.Message != "" {
			extra = fmt.Sprintf("%s - %s", extra, payload.Message)
		}
	} else {
		extra = fmt.Sprintf("unexpected status code: %d", e.Resp.StatusCode)
	}

	return fmt.Sprintf("Could not %s %s '%s', %s", e.action, e.resource, e.resourceID, extra)
}

// errorPayload parses the error payload sent by the api client from the body buffered at construction.
// It returns nil if there is no payload.
func (e APIError) errorPayload() *apiErrorPayload {
	if e.err == nil || len(e.bufferedBody) == 0 {
		return nil
	}

	var payload apiErrorPayload
	if err := json.Unmarshal(e.bufferedBody, &payload); err != nil {
		return nil
	}

	return &payload
}

// NewAPIErrorFromError finds an APIError in err's wrap chain.
// This is useful when working with an APIError passed around as an `error`, possibly wrapped by service layers.
// It returns nil when the chain contains no APIError.
func NewAPIErrorFromError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

// IsErrNotFound reports whether err's wrap chain contains an APIError that is a 404 (per IsNotFound).
func IsErrNotFound(err error) bool {
	apiErr := NewAPIErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsNotFound()
}

// IsErrBadRequest reports whether err's wrap chain contains an APIError that is a 400 (per IsBadRequest).
func IsErrBadRequest(err error) bool {
	apiErr := NewAPIErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsBadRequest()
}

// NewAPIError returns a new instance of APIError with the given parameters.
// The response body, when present, is buffered here so the error can be inspected multiple
// times (e.g. IsNotFound then Error) without draining the stream.
func NewAPIError(action APIAction, resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	var bufferedBody []byte
	if resp != nil && resp.Body != nil {
		if body, readErr := io.ReadAll(resp.Body); readErr == nil {
			bufferedBody = body
		}
		resp.Body.Close()
	}

	return &APIError{
		err:          err,
		action:       action,
		resource:     resource,
		resourceID:   resourceID,
		Resp:         resp,
		bufferedBody: bufferedBody,
	}
}

// NewCreateAPIError returns a new instance of APIError for a `create` action with the given parameters.
func NewCreateAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionCreate, resource, resourceID, resp, err)
}

// NewReadAPIError returns a new instance of APIError for a `read` action with the given parameters.
func NewReadAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionRead, resource, resourceID, resp, err)
}

// NewUpdateAPIError returns a new instance of APIError for an `update` action with the given parameters.
func NewUpdateAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionUpdate, resource, resourceID, resp, err)
}

// NewDeleteAPIError returns a new instance of APIError for a `delete` action with the given parameters.
func NewDeleteAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionDelete, resource, resourceID, resp, err)
}

// NewStopAPIError returns a new instance of APIError for a `stop` action with the given parameters.
func NewStopAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionStop, resource, resourceID, resp, err)
}

// NewRedeployAPIError returns a new instance of APIError for a `redeploy` action with the given parameters.
func NewRedeployAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionRedeploy, resource, resourceID, resp, err)
}

// NewDeployAPIError returns a new instance of APIError for a `deploy` action with the given parameters.
func NewDeployAPIError(resource APIResource, resourceID string, resp *http.Response, err error) *APIError {
	return NewAPIError(APIActionDeploy, resource, resourceID, resp, err)
}

// NewNotFoundAPIError returns a new instance of APIError for a `not_found` resource with the given parameters.
func NewNotFoundAPIError(resource APIResource, resourceID string) *APIError {
	return NewAPIError(APIActionRead, resource, resourceID, &http.Response{
		StatusCode: http.StatusNotFound,
	}, errors.New("resource not found"))
}
