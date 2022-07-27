package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// apiAction is an enum that contains every type of actions done using the api.
// This is used to generate a detailed error message displayed by terraform when the api return an error.
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

// apiResource is an enum that contains every resource we handle using the api .
// This is used to generate a detailed error message displayed by terraform when the api return an error.
type apiResource string

const (
	apiResourceAWSCredentials                 apiResource = "aws credentials"
	apiResourceApplication                    apiResource = "application"
	apiResourceApplicationCustomDomain        apiResource = "application custom domain"
	apiResourceApplicationEnvironmentVariable apiResource = "application environment variable"
	apiResourceApplicationSecret              apiResource = "application secret"
	apiResourceApplicationStatus              apiResource = "application status"
	apiResourceCluster                        apiResource = "cluster"
	apiResourceClusterCloudProvider           apiResource = "cluster cloud provider"
	apiResourceClusterInstanceType            apiResource = "cluster instance type"
	apiResourceClusterRoutingTable            apiResource = "cluster routing table"
	apiResourceClusterStatus                  apiResource = "cluster status"
	apiResourceDatabase                       apiResource = "database"
	apiResourceDatabaseStatus                 apiResource = "database status"
	apiResourceEnvironment                    apiResource = "environment"
	apiResourceEnvironmentEnvironmentVariable apiResource = "environment environment variable"
	apiResourceEnvironmentSecret              apiResource = "environment secret"
	apiResourceEnvironmentStatus              apiResource = "environment status"
	apiResourceOrganization                   apiResource = "organization"
	apiResourceProject                        apiResource = "project"
	apiResourceProjectEnvironmentVariable     apiResource = "project environment variable"
	apiResourceProjectSecret                  apiResource = "project secret"
	apiResourceScalewayCredentials            apiResource = "scaleway credentials"
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

// apiError represents an error that comes from Qovery's API client.
// It contains all the information needed to understand an api error.
type apiError struct {
	err        error          // err is the actual error returned by the api client.
	action     apiAction      // action that produced the api client error.
	resource   apiResource    // resource that produced the api client error.
	resourceID string         // resourceID that produced the api client error. [NOTE: it is replaced by the resource name in some cases (for failed create requests)]
	resp       *http.Response // resp is the response returned by the api client.
}

// IsNotFound returns weather the error is a 404 or not.
// NOTE: Since the api returns a 403 when a resource is not found, we also consider those as a 404.
func (e apiError) IsNotFound() bool {
	if e.resp == nil {
		return false
	}

	// NOTE: consider 403 Forbidden as a 404 NotFound until the api is fixed
	return e.resp.StatusCode == http.StatusNotFound ||
		e.resp.StatusCode == http.StatusForbidden
}

// IsBadRequest returns weather the error is a 400 or not.
func (e apiError) IsBadRequest() bool {
	if e.resp == nil {
		return false
	}

	return e.resp.StatusCode == http.StatusBadRequest
}

// Error implements the Error interface.
// It returns the detailed error message for this apiError.
func (e apiError) Error() string {
	return e.Detail()
}

// Summary returns a brief description of the error with the apiResource and apiAction.
func (e apiError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

// Detail return a detailed error message for this apiError.
// It tries to read the error payload received from the api client to give extra information about the error.
func (e apiError) Detail() string {
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

// newApiErrorFromError tries to cast an error into an apiError.
// This is useful when working with apiError passed as an `error` type to get the actual apiError type.
func newApiErrorFromError(err error) *apiError {
	switch err.(type) {
	case *apiError:
		return err.(*apiError)
	default:
		return nil
	}
}

// IsErrNotFound takes an error type and tries to cast it into a apiError to check weather the error is an 404 or not.
// It returns false if the casting fails.
func IsErrNotFound(err error) bool {
	apiErr := newApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsNotFound()
}

// IsErrBadRequest takes an error type and tries to cast it into a apiError to check weather the error is an 400 or not.
// It returns false if the casting fails.
func IsErrBadRequest(err error) bool {
	apiErr := newApiErrorFromError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.IsBadRequest()
}

// newApiError returns a new instance of apiError with the given parameters.
func newApiError(action apiAction, resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return &apiError{
		err:        err,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		resp:       resp,
	}
}

// newCreateApiError returns a new instance of apiError for a `create` action with the given parameters.
func newCreateApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionCreate, resource, resourceID, resp, err)
}

// newReadApiError returns a new instance of apiError for a `read` action with the given parameters.
func newReadApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionRead, resource, resourceID, resp, err)
}

// newUpdateApiError returns a new instance of apiError for an `update` action with the given parameters.
func newUpdateApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionUpdate, resource, resourceID, resp, err)
}

// newDeleteApiError returns a new instance of apiError for a `delete` action with the given parameters.
func newDeleteApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionDelete, resource, resourceID, resp, err)
}

// newStopApiError returns a new instance of apiError for a `stop` action with the given parameters.
func newStopApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionStop, resource, resourceID, resp, err)
}

// newRestartApiError returns a new instance of apiError for a `restart` action with the given parameters.
func newRestartApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionRestart, resource, resourceID, resp, err)
}

// newDeployApiError returns a new instance of apiError for a `deploy` action with the given parameters.
func newDeployApiError(resource apiResource, resourceID string, resp *http.Response, err error) *apiError {
	return newApiError(apiActionDeploy, resource, resourceID, resp, err)
}
