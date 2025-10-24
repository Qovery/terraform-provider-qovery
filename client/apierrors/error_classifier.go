package apierrors

import (
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
)

// IsTransientError checks if an error is temporary and may succeed on retry
func IsTransientError(apiErr *APIError) bool {
	if apiErr == nil {
		return false
	}

	// Check HTTP status codes for transient errors
	if apiErr.res != nil {
		statusCode := apiErr.res.StatusCode

		// Server errors (5xx) are typically transient
		if statusCode >= 500 && statusCode < 600 {
			return true
		}

		// Rate limiting - should retry with backoff
		if statusCode == http.StatusTooManyRequests {
			return true
		}
	}

	// Check underlying error for network-related issues
	if apiErr.err != nil {
		return isNetworkError(apiErr.err)
	}

	return false
}

// isNetworkError checks if an error is a network/connection error
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for EOF errors
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Check for network timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for connection refused/reset errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check for syscall errors (connection reset, broken pipe, etc.)
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	return false
}

// IsRetryable determines if an API error should be retried
func IsRetryable(apiErr *APIError) bool {
	if apiErr == nil {
		return false
	}

	// Don't retry client errors (4xx) except rate limiting
	if apiErr.res != nil {
		statusCode := apiErr.res.StatusCode
		if statusCode >= 400 && statusCode < 500 {
			// Only retry rate limiting
			return statusCode == http.StatusTooManyRequests
		}
	}

	// Retry transient errors
	return IsTransientError(apiErr)
}
