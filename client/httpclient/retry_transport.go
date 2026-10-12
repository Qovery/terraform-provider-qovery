// Package httpclient builds the HTTP client every Qovery API call of the provider goes through.
package httpclient

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// Retry policy defaults of the client returned by New.
const (
	DefaultMaxAttempts    = 3
	DefaultInitialBackoff = 2 * time.Second
	DefaultMaxBackoff     = 30 * time.Second
	DefaultMaxRetryAfter  = 60 * time.Second
)

// New returns an HTTP client that retries transient Qovery API failures with RetryTransport.
func New() *http.Client {
	return &http.Client{Transport: &RetryTransport{}}
}

// RetryTransport is an http.RoundTripper that retries transient failures with exponential
// backoff and jitter. It honours Retry-After and stops waiting when the request context ends.
//
// What it retries depends on whether replaying the request is safe:
//   - GET, HEAD, OPTIONS and DELETE: 429, 5xx and network errors, a connection lost while
//     reading the response body included. Their response body is read before it is returned.
//   - POST, PUT and PATCH: 429, and a connection that could not be established. A 5xx, or a
//     connection lost once the request was sent, may hide a create or a deploy the API already
//     performed, so the response or error is returned as is.
//
// A request whose body cannot be rewound (no GetBody) is never retried.
//
// Whenever it stops, because an attempt succeeded, is not retried, used the last attempt, or
// the context ended during the backoff, it returns the outcome of the last attempt unchanged:
// its response, or its error when no response was received.
//
// A zero field falls back to its default.
type RetryTransport struct {
	// Base sends each attempt. Defaults to http.DefaultTransport.
	Base http.RoundTripper
	// MaxAttempts is the total number of attempts, the first one included.
	MaxAttempts int
	// InitialBackoff is the wait before the first retry. It doubles at each retry.
	InitialBackoff time.Duration
	// MaxBackoff caps the computed backoff.
	MaxBackoff time.Duration
	// MaxRetryAfter is the longest Retry-After the transport waits for. A response asking to
	// wait longer is returned without a retry.
	MaxRetryAfter time.Duration
}

// RoundTrip implements http.RoundTripper.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	maxAttempts := valueOrDefault(t.MaxAttempts, DefaultMaxAttempts)
	backoff := valueOrDefault(t.InitialBackoff, DefaultInitialBackoff)
	maxBackoff := valueOrDefault(t.MaxBackoff, DefaultMaxBackoff)
	maxRetryAfter := valueOrDefault(t.MaxRetryAfter, DefaultMaxRetryAfter)

	for attempt := 1; ; attempt++ {
		attemptReq, err := requestForAttempt(req, attempt)
		if err != nil {
			return nil, err
		}

		resp, err := t.base().RoundTrip(attemptReq)
		retry, failure := evaluateAttempt(req.Method, resp, err)
		if !retry || attempt >= maxAttempts || ctx.Err() != nil || !canReplay(req) {
			return resp, err
		}

		delay := applyJitter(backoff)
		if retryAfter, ok := parseRetryAfter(resp, time.Now()); ok {
			if retryAfter > maxRetryAfter {
				return resp, err
			}
			delay = max(delay, retryAfter)
		}

		tflog.Warn(ctx, "Retrying a Qovery API request after a transient failure", map[string]any{
			"method":       req.Method,
			"path":         req.URL.Path,
			"failure":      failure,
			"next_attempt": attempt + 1,
			"max_attempts": maxAttempts,
			"delay":        delay.String(),
		})

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return resp, err
		case <-timer.C:
		}

		if resp != nil {
			_ = resp.Body.Close()
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (t *RetryTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func valueOrDefault[T int | time.Duration](value, defaultValue T) T {
	if value > 0 {
		return value
	}
	return defaultValue
}

// requestForAttempt returns the request to send for the given attempt. The first attempt sends
// req itself; a retry sends a clone with a fresh body, since the previous attempt consumed it.
func requestForAttempt(req *http.Request, attempt int) (*http.Request, error) {
	if attempt == 1 {
		return req, nil
	}

	retryReq := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		retryReq.Body = body
	}
	return retryReq, nil
}

// canReplay reports whether req can be sent again: it has no body, or a body GetBody can rewind.
func canReplay(req *http.Request) bool {
	return req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
}

// evaluateAttempt reports whether the outcome of an attempt is retried, and describes its failure.
//
// It reads into memory the body of every idempotent response and of every response it retries.
// A connection lost while reading the body of an idempotent response is then a retryable failure,
// and a retried response frees its connection during the backoff yet can still be returned.
func evaluateAttempt(method string, resp *http.Response, err error) (bool, string) {
	if err != nil {
		if isIdempotent(method) {
			return apierrors.IsNetworkError(err), err.Error()
		}
		return isDialError(err), err.Error()
	}

	retry := shouldRetryStatus(method, resp.StatusCode)
	if !retry && !isIdempotent(method) {
		return false, ""
	}
	if bodyErr := bufferBody(resp); bodyErr != nil && isIdempotent(method) && apierrors.IsNetworkError(bodyErr) {
		return true, bodyErr.Error()
	}
	return retry, resp.Status
}

func shouldRetryStatus(method string, statusCode int) bool {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return true
	case statusCode >= 500:
		return isIdempotent(method)
	default:
		return false
	}
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		return true
	default:
		return false
	}
}

// isDialError reports whether err happened while opening the connection, DNS lookup included,
// so the request never reached the server.
func isDialError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}

// parseRetryAfter returns the wait a response asks for in its Retry-After header, given either
// as seconds or as an HTTP date.
func parseRetryAfter(resp *http.Response, now time.Time) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	value := resp.Header.Get("Retry-After")
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if date, err := http.ParseTime(value); err == nil {
		return max(date.Sub(now), 0), true
	}
	return 0, false
}

// applyJitter returns a random duration between backoff/2 and backoff, so that concurrent
// clients do not retry in lockstep.
func applyJitter(backoff time.Duration) time.Duration {
	half := backoff / 2
	if half <= 0 {
		return backoff
	}
	return half + time.Duration(rand.Int63n(int64(half)))
}

// bufferBody reads the whole response body and replaces it with an in-memory copy, closing the
// connection body. When the read fails, the copy holds the bytes read followed by the read
// error, so the caller's own read fails the same way.
func bufferBody(resp *http.Response) error {
	if resp.Body == nil || resp.Body == http.NoBody {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var copied io.Reader = bytes.NewReader(body)
	if err != nil {
		copied = io.MultiReader(copied, errorReader{err: err})
	}
	resp.Body = io.NopCloser(copied)
	return err
}

// errorReader fails every read with err.
type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}
