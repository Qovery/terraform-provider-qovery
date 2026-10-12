//go:build unit && !integration
// +build unit,!integration

package client

import (
	"net/http"
	"time"

	"github.com/qovery/terraform-provider-qovery/client/httpclient"
)

// newTestHTTPClient returns a client with the production retry policy and millisecond backoffs,
// so tests exercise the retries without waiting for them.
func newTestHTTPClient() *http.Client {
	return &http.Client{Transport: &httpclient.RetryTransport{
		InitialBackoff: time.Millisecond,
		MaxBackoff:     time.Millisecond,
	}}
}
