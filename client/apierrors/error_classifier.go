package apierrors

import (
	"errors"
	"io"
	"net"
	"syscall"
)

// IsNetworkError reports whether err is a network or connection error, which may succeed on retry.
func IsNetworkError(err error) bool {
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
