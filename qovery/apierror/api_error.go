package apierror

import (
	"fmt"
	"net/http"
)

type APIAction string

var (
	Create APIAction = "create"
	Read   APIAction = "read"
	Update APIAction = "update"
	Delete APIAction = "delete"
	Deploy APIAction = "deploy"
)

type APIError struct {
	resource   string
	resourceID string
	action     APIAction
	res        *http.Response
	err        error
}

func New(resource string, resourceID string, action APIAction, res *http.Response, err error) *APIError {
	return &APIError{
		resource:   resource,
		resourceID: resourceID,
		action:     action,
		res:        res,
		err:        err,
	}
}

func (e APIError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

func (e APIError) Detail() string {
	var extra string
	if e.err != nil {
		extra = fmt.Sprintf("unexpected error: %s", e.err)
	} else {
		extra = fmt.Sprintf("unexpected status code: %d", e.res.StatusCode)
	}
	return fmt.Sprintf("Could not %s %s '%s', %s", e.action, e.resource, e.resourceID, extra)
}
