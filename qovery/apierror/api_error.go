package apierror

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIAction string

var (
	Create  APIAction = "create"
	Read    APIAction = "read"
	Update  APIAction = "update"
	Delete  APIAction = "delete"
	Deploy  APIAction = "deploy"
	Restart APIAction = "restart"
)

type ErrorPayload struct {
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

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
	payload := e.ErrorPayload()

	if e.err != nil {
		extra = fmt.Sprintf("unexpected error: %s", e.err)
		if payload != nil && payload.Message != "" {
			extra = fmt.Sprintf("unexpected error: %s - %s", e.err, payload.Message)
		}
	} else {
		extra = fmt.Sprintf("unexpected status code: %d", e.res.StatusCode)
	}
	return fmt.Sprintf("Could not %s %s '%s', %s", e.action, e.resource, e.resourceID, extra)
}

func (e APIError) ErrorPayload() *ErrorPayload {
	if e.err == nil {
		return nil
	}

	body, err := io.ReadAll(e.res.Body)
	if err != nil {
		return nil
	}

	var payload ErrorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	return &payload
}
