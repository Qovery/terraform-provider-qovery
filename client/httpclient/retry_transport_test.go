//go:build unit && !integration
// +build unit,!integration

package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fastTransport returns a RetryTransport with the default attempts and millisecond backoffs.
func fastTransport(base http.RoundTripper) *RetryTransport {
	return &RetryTransport{Base: base, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond}
}

// scriptedServer answers the n-th request with the n-th status, repeating the last status once
// the script runs out. A 429 carries retryAfter in its Retry-After header when set. It records
// the body of every request and writes the attempt number in each response body.
type scriptedServer struct {
	mu         sync.Mutex
	statuses   []int
	retryAfter string
	bodies     []string
}

func (s *scriptedServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	body, _ := io.ReadAll(r.Body)
	s.bodies = append(s.bodies, string(body))
	attempt := len(s.bodies)
	status := s.statuses[min(attempt, len(s.statuses))-1]

	if status == http.StatusTooManyRequests && s.retryAfter != "" {
		w.Header().Set("Retry-After", s.retryAfter)
	}
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"attempt":%d}`, attempt)
}

func (s *scriptedServer) recordedBodies() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...)
}

func newRequest(t *testing.T, ctx context.Context, method, url, body string) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	require.NoError(t, err)
	return req
}

func TestRetryTransport_RetriesByMethodAndStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName     string
		Method       string
		Body         string
		Statuses     []int
		ExpectCalls  int
		ExpectStatus int
	}{
		{TestName: "get_success_is_sent_once", Method: http.MethodGet, Statuses: []int{200}, ExpectCalls: 1, ExpectStatus: 200},
		{TestName: "get_429_then_200_is_retried", Method: http.MethodGet, Statuses: []int{429, 200}, ExpectCalls: 2, ExpectStatus: 200},
		{TestName: "get_503_then_200_is_retried", Method: http.MethodGet, Statuses: []int{503, 200}, ExpectCalls: 2, ExpectStatus: 200},
		{TestName: "get_500_on_every_attempt_returns_the_last_response", Method: http.MethodGet, Statuses: []int{500}, ExpectCalls: DefaultMaxAttempts, ExpectStatus: 500},
		{TestName: "get_404_is_not_retried", Method: http.MethodGet, Statuses: []int{404, 200}, ExpectCalls: 1, ExpectStatus: 404},
		{TestName: "delete_500_then_200_is_retried", Method: http.MethodDelete, Statuses: []int{500, 200}, ExpectCalls: 2, ExpectStatus: 200},
		{TestName: "post_500_is_not_retried", Method: http.MethodPost, Body: `{"name":"app"}`, Statuses: []int{500, 201}, ExpectCalls: 1, ExpectStatus: 500},
		{TestName: "put_502_is_not_retried", Method: http.MethodPut, Body: `{"name":"app"}`, Statuses: []int{502, 200}, ExpectCalls: 1, ExpectStatus: 502},
		{TestName: "patch_503_is_not_retried", Method: http.MethodPatch, Body: `{"name":"app"}`, Statuses: []int{503, 200}, ExpectCalls: 1, ExpectStatus: 503},
		{TestName: "post_429_then_201_is_retried_with_the_same_body", Method: http.MethodPost, Body: `{"name":"app"}`, Statuses: []int{429, 201}, ExpectCalls: 2, ExpectStatus: 201},
		{TestName: "put_429_then_200_is_retried_with_the_same_body", Method: http.MethodPut, Body: `{"name":"app"}`, Statuses: []int{429, 200}, ExpectCalls: 2, ExpectStatus: 200},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &scriptedServer{statuses: tc.Statuses}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			httpClient := &http.Client{Transport: fastTransport(nil)}
			resp, err := httpClient.Do(newRequest(t, context.Background(), tc.Method, server.URL, tc.Body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.ExpectStatus, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf(`{"attempt":%d}`, tc.ExpectCalls), string(body), "the caller gets the last response, body intact")

			bodies := api.recordedBodies()
			require.Len(t, bodies, tc.ExpectCalls)
			for _, received := range bodies {
				assert.Equal(t, tc.Body, received, "every attempt sends the full request body")
			}
		})
	}
}

func TestRetryTransport_RetryAfter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		RetryAfter    string
		MaxRetryAfter time.Duration
		ExpectCalls   int
		ExpectStatus  int
		ExpectMinWait time.Duration
		ExpectMaxWait time.Duration
	}{
		{
			TestName:      "waits_for_retry_after_seconds",
			RetryAfter:    "1",
			ExpectCalls:   2,
			ExpectStatus:  200,
			ExpectMinWait: time.Second,
			ExpectMaxWait: 5 * time.Second,
		},
		{
			TestName:      "returns_the_429_when_retry_after_exceeds_the_cap",
			RetryAfter:    "120",
			MaxRetryAfter: time.Second,
			ExpectCalls:   1,
			ExpectStatus:  429,
			// Far below the 120s asked for: it proves the wait was skipped, with room for a
			// slow test machine.
			ExpectMaxWait: 30 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &scriptedServer{statuses: []int{429, 200}, retryAfter: tc.RetryAfter}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			transport := fastTransport(nil)
			transport.MaxRetryAfter = tc.MaxRetryAfter
			httpClient := &http.Client{Transport: transport}

			start := time.Now()
			resp, err := httpClient.Do(newRequest(t, context.Background(), http.MethodGet, server.URL, ""))
			elapsed := time.Since(start)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.ExpectStatus, resp.StatusCode)
			assert.Len(t, api.recordedBodies(), tc.ExpectCalls)
			assert.GreaterOrEqual(t, elapsed, tc.ExpectMinWait)
			assert.Less(t, elapsed, tc.ExpectMaxWait)
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	testCases := []struct {
		TestName    string
		Header      string
		ExpectWait  time.Duration
		ExpectFound bool
	}{
		{TestName: "missing_header", Header: "", ExpectFound: false},
		{TestName: "seconds", Header: "3", ExpectWait: 3 * time.Second, ExpectFound: true},
		{TestName: "zero_seconds", Header: "0", ExpectWait: 0, ExpectFound: true},
		{TestName: "negative_seconds", Header: "-1", ExpectFound: false},
		{TestName: "http_date_in_the_future", Header: now.Add(10 * time.Second).Format(http.TimeFormat), ExpectWait: 10 * time.Second, ExpectFound: true},
		{TestName: "http_date_in_the_past", Header: now.Add(-time.Minute).Format(http.TimeFormat), ExpectWait: 0, ExpectFound: true},
		{TestName: "garbage", Header: "soon", ExpectFound: false},
		{TestName: "seconds_beyond_a_duration_are_clamped", Header: "99999999999999", ExpectWait: time.Duration(math.MaxInt64), ExpectFound: true},
		{TestName: "seconds_beyond_int64_are_clamped", Header: "99999999999999999999999", ExpectWait: time.Duration(math.MaxInt64), ExpectFound: true},
		{TestName: "seconds_below_int64_are_ignored", Header: "-99999999999999999999999", ExpectFound: false},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{Header: http.Header{}}
			if tc.Header != "" {
				resp.Header.Set("Retry-After", tc.Header)
			}

			wait, found := parseRetryAfter(resp, now)
			assert.Equal(t, tc.ExpectFound, found)
			assert.Equal(t, tc.ExpectWait, wait)
		})
	}

	t.Run("nil_response", func(t *testing.T) {
		t.Parallel()
		_, found := parseRetryAfter(nil, now)
		assert.False(t, found)
	})
}

// TestRetryTransport_ReturnsTheLastResponseWhenTheContextEnds pins that cancelling the context
// stops the backoff and hands back the last response as is, as if the request had not been
// retried. Whether the cancellation lands before or during the backoff, the outcome is the same.
func TestRetryTransport_ReturnsTheLastResponseWhenTheContextEnds(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	base := &sequenceTransport{responses: []*http.Response{{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 Service Unavailable",
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"message":"unavailable"}`)),
	}}}
	transport := &RetryTransport{Base: base, InitialBackoff: time.Minute, MaxBackoff: time.Minute}

	start := time.Now()
	resp, err := transport.RoundTrip(newRequest(t, ctx, http.MethodGet, "https://api.qovery.test/organization", ""))

	require.NoError(t, err)
	assert.Less(t, time.Since(start), 5*time.Second, "the backoff must not outlive the context")
	assert.Equal(t, 1, base.calls)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"message":"unavailable"}`, string(body))
}

func TestRetryTransport_RequestWithoutGetBodyIsNotRetried(t *testing.T) {
	t.Parallel()

	api := &scriptedServer{statuses: []int{429, 200}}
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)

	req := newRequest(t, context.Background(), http.MethodPut, server.URL, "")
	req.Body = io.NopCloser(strings.NewReader(`{"name":"app"}`))
	req.GetBody = nil

	resp, err := (&http.Client{Transport: fastTransport(nil)}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Len(t, api.recordedBodies(), 1)
}

// TestRetryTransport_ReturnsTheLastResponseWhenTheBodyCannotBeRewound pins that a retry whose
// request body cannot be rewound is not sent, and the last response is returned as is.
func TestRetryTransport_ReturnsTheLastResponseWhenTheBodyCannotBeRewound(t *testing.T) {
	t.Parallel()

	base := &sequenceTransport{responses: []*http.Response{{
		StatusCode: http.StatusTooManyRequests,
		Status:     "429 Too Many Requests",
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"message":"slow down"}`)),
	}}}
	req := newRequest(t, context.Background(), http.MethodPut, "https://api.qovery.test/organization", `{"name":"org"}`)
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, errors.New("rewind failed")
	}

	resp, err := fastTransport(base).RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, 1, base.calls)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"message":"slow down"}`, string(body))
}

// TestRetryTransport_AbandonedRetryDoesNotRewindTheBody pins that the request body is rewound
// only for a retry that is sent, so an abandoned retry leaves no rewound body open.
func TestRetryTransport_AbandonedRetryDoesNotRewindTheBody(t *testing.T) {
	t.Parallel()

	base := &sequenceTransport{responses: []*http.Response{{
		StatusCode: http.StatusTooManyRequests,
		Status:     "429 Too Many Requests",
		Header:     http.Header{"Retry-After": []string{"120"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}}}
	req := newRequest(t, context.Background(), http.MethodPut, "https://api.qovery.test/organization", `{"name":"org"}`)
	rewinds := 0
	req.GetBody = func() (io.ReadCloser, error) {
		rewinds++
		return io.NopCloser(strings.NewReader(`{"name":"org"}`)), nil
	}

	transport := fastTransport(base)
	transport.MaxRetryAfter = time.Second
	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, 1, base.calls)
	assert.Zero(t, rewinds, "a retry abandoned over Retry-After must not rewind the body")
}

// scriptedTransport fails the first calls with errs, then answers 200.
type scriptedTransport struct {
	mu    sync.Mutex
	errs  []error
	calls int
}

func (s *scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++
	if s.calls <= len(s.errs) {
		return nil, s.errs[s.calls-1]
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

func TestRetryTransport_ConnectionErrors(t *testing.T) {
	t.Parallel()

	dialRefused := &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}
	dnsFailure := &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "no such host", Name: "api.qovery.com", IsNotFound: true}}
	resetAfterWrite := &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}

	testCases := []struct {
		TestName    string
		Method      string
		Err         error
		ExpectCalls int
		ExpectError bool
	}{
		{TestName: "get_eof_is_retried", Method: http.MethodGet, Err: io.EOF, ExpectCalls: 2},
		{TestName: "get_reset_is_retried", Method: http.MethodGet, Err: resetAfterWrite, ExpectCalls: 2},
		{TestName: "delete_eof_is_retried", Method: http.MethodDelete, Err: io.EOF, ExpectCalls: 2},
		{TestName: "post_refused_connection_is_retried", Method: http.MethodPost, Err: dialRefused, ExpectCalls: 2},
		{TestName: "post_dns_failure_is_retried", Method: http.MethodPost, Err: dnsFailure, ExpectCalls: 2},
		{TestName: "put_refused_connection_is_retried", Method: http.MethodPut, Err: dialRefused, ExpectCalls: 2},
		{TestName: "post_eof_is_not_retried", Method: http.MethodPost, Err: io.EOF, ExpectCalls: 1, ExpectError: true},
		{TestName: "post_reset_after_write_is_not_retried", Method: http.MethodPost, Err: resetAfterWrite, ExpectCalls: 1, ExpectError: true},
		{TestName: "get_non_network_error_is_not_retried", Method: http.MethodGet, Err: errors.New("tls: bad certificate"), ExpectCalls: 1, ExpectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			base := &scriptedTransport{errs: []error{tc.Err}}
			req := newRequest(t, context.Background(), tc.Method, "https://api.qovery.test/organization", `{"name":"org"}`)
			resp, err := fastTransport(base).RoundTrip(req)

			assert.Equal(t, tc.ExpectCalls, base.calls)
			if tc.ExpectError {
				assert.ErrorIs(t, err, tc.Err)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// countingTransport counts the attempts it forwards to http.DefaultTransport.
type countingTransport struct {
	mu    sync.Mutex
	calls int
}

func (c *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

// TestRetryTransport_PostToAClosedServerIsRetried pins that Go reports a refused connection as a
// dial error, the only network error a POST is retried on.
func TestRetryTransport_PostToAClosedServerIsRetried(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL
	server.Close()

	base := &countingTransport{}
	resp, err := (&http.Client{Transport: fastTransport(base)}).Do(newRequest(t, context.Background(), http.MethodPost, url, `{"name":"app"}`))
	if resp != nil {
		resp.Body.Close()
	}

	require.Error(t, err)
	assert.True(t, isDialError(err), "a refused connection must be a dial error, got %v", err)
	assert.Equal(t, DefaultMaxAttempts, base.calls)
}

// trackedBody records whether it was read to the end and closed.
type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

// sequenceTransport answers with the given responses in order.
type sequenceTransport struct {
	responses []*http.Response
	calls     int
}

func (s *sequenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := s.responses[s.calls]
	s.calls++
	resp.Request = req
	return resp, nil
}

func TestRetryTransport_ReadsAndClosesRetriedResponses(t *testing.T) {
	t.Parallel()

	discardedReader := strings.NewReader(`{"message":"unavailable"}`)
	discarded := &trackedBody{Reader: discardedReader}
	final := &trackedBody{Reader: strings.NewReader(`{"id":"1"}`)}
	base := &sequenceTransport{responses: []*http.Response{
		{StatusCode: http.StatusTooManyRequests, Status: "429 Too Many Requests", Header: http.Header{}, Body: discarded},
		{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{}, Body: final},
	}}

	// The transport reads the body of a PUT response only when it retries it, so the final 200
	// reaches the caller untouched.
	resp, err := fastTransport(base).RoundTrip(newRequest(t, context.Background(), http.MethodPut, "https://api.qovery.test/organization", `{"name":"org"}`))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, discarded.closed, "the retried response must be closed")
	assert.Zero(t, discardedReader.Len(), "the retried response must be read so its connection is reused")
	assert.False(t, final.closed, "the returned response belongs to the caller")
}

// truncatingServer cuts the connection in the middle of the response body of the first
// truncations requests, then answers the next requests in full.
type truncatingServer struct {
	mu          sync.Mutex
	truncations int
	calls       int
}

const truncatedBody = `{"status":"DEPLOYED"}`

func (s *truncatingServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	s.calls++
	call := s.calls
	s.mu.Unlock()

	if call <= s.truncations {
		w.Header().Set("Content-Length", strconv.Itoa(len(truncatedBody)+100))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(truncatedBody))
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	}
	_, _ = w.Write([]byte(truncatedBody))
}

func (s *truncatingServer) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func TestRetryTransport_ConnectionLostWhileReadingTheBody(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		Method          string
		Truncations     int
		ExpectCalls     int
		ExpectReadError bool
	}{
		{TestName: "get_is_retried", Method: http.MethodGet, Truncations: 1, ExpectCalls: 2},
		{TestName: "delete_is_retried", Method: http.MethodDelete, Truncations: 1, ExpectCalls: 2},
		{TestName: "post_is_not_retried", Method: http.MethodPost, Truncations: 1, ExpectCalls: 1, ExpectReadError: true},
		// Once the attempts are used up, the caller gets the response and a failing read, as
		// it would without the transport.
		{TestName: "get_truncated_on_every_attempt_returns_the_response", Method: http.MethodGet, Truncations: DefaultMaxAttempts, ExpectCalls: DefaultMaxAttempts, ExpectReadError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &truncatingServer{truncations: tc.Truncations}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			resp, err := (&http.Client{Transport: fastTransport(nil)}).Do(newRequest(t, context.Background(), tc.Method, server.URL, ""))
			require.NoError(t, err)
			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)

			assert.Equal(t, tc.ExpectCalls, api.callCount())
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			if tc.ExpectReadError {
				assert.ErrorIs(t, readErr, io.ErrUnexpectedEOF)
				assert.Equal(t, truncatedBody, string(body), "the bytes received before the cut are kept")
				return
			}
			require.NoError(t, readErr)
			assert.Equal(t, truncatedBody, string(body))
		})
	}
}

func TestRetryTransport_BuffersAndClosesSuccessfulIdempotentResponses(t *testing.T) {
	t.Parallel()

	original := &trackedBody{Reader: strings.NewReader(`{"id":"1"}`)}
	base := &sequenceTransport{responses: []*http.Response{
		{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{}, Body: original},
	}}

	resp, err := fastTransport(base).RoundTrip(newRequest(t, context.Background(), http.MethodGet, "https://api.qovery.test/organization", ""))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.True(t, original.closed, "the connection body is closed once read into memory")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"id":"1"}`, string(body))
}

func TestApplyJitter(t *testing.T) {
	t.Parallel()

	assert.Zero(t, applyJitter(0))

	for _, backoff := range []time.Duration{time.Nanosecond, 10 * time.Millisecond, 2 * time.Second, 30 * time.Second} {
		seen := map[time.Duration]struct{}{}
		for range 50 {
			jittered := applyJitter(backoff)
			assert.GreaterOrEqual(t, jittered, backoff/2)
			assert.LessOrEqual(t, jittered, backoff)
			seen[jittered] = struct{}{}
		}
		if backoff >= 10*time.Millisecond {
			assert.Greater(t, len(seen), 1, "jitter must spread the retries of %s", backoff)
		}
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	httpClient := New()
	require.IsType(t, &RetryTransport{}, httpClient.Transport)
	assert.Equal(t, http.DefaultTransport, httpClient.Transport.(*RetryTransport).base(), "connections are pooled like http.DefaultClient")
}
