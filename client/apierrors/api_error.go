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
	return e.res.StatusCode == http.StatusNotFound
}

type errorPayload struct {
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
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
	if e.err == nil {
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
