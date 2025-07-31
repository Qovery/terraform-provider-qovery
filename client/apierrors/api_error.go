package apierrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	err        error
	action     APIAction
	resource   APIResource
	resourceID string
	res        *http.Response
}

func IsNotFound(e *APIError) bool {
	if e == nil || e.res == nil {
		return false
	}

	// NOTE: consider 403 Forbidden as a 404 NotFound until the api is fixed
	return e.res.StatusCode == http.StatusNotFound ||
		e.res.StatusCode == http.StatusForbidden
}

func IsBadRequest(e *APIError) bool {
	if e == nil || e.res == nil {
		return false
	}

	return e.res.StatusCode == http.StatusBadRequest
}

type errorPayload struct {
	Status    int    `json:"status"`
	Message   string `json:"detail"`
}

func (e APIError) Error() string {
	return e.Detail()
}

func (e APIError) Summary() string {
	return fmt.Sprintf("Error on %s %s", e.resource, e.action)
}

func (e APIError) Detail() string {
	var extra string
	payload := e.errorPayload()

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

func (e APIError) errorPayload() *errorPayload {
	if e.err == nil || e.res == nil {
		return nil
	}

	body, err := io.ReadAll(e.res.Body)
	if err != nil {
		return nil
	}

	var payload errorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	return &payload
}
