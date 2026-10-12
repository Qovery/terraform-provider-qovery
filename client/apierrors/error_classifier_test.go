//go:build unit && !integration

package apierrors

import (
	"errors"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsNetworkError verifies network error detection
func TestIsNetworkError(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		isNetwork bool
	}{
		{
			name:      "should return false for nil error",
			err:       nil,
			isNetwork: false,
		},
		{
			name:      "should detect EOF as network error",
			err:       io.EOF,
			isNetwork: true,
		},
		{
			name:      "should detect unexpected EOF as network error",
			err:       io.ErrUnexpectedEOF,
			isNetwork: true,
		},
		{
			name:      "should detect connection refused as network error",
			err:       syscall.ECONNREFUSED,
			isNetwork: true,
		},
		{
			name:      "should detect connection reset as network error",
			err:       syscall.ECONNRESET,
			isNetwork: true,
		},
		{
			name:      "should detect broken pipe as network error",
			err:       syscall.EPIPE,
			isNetwork: true,
		},
		{
			name: "should detect net.OpError as network error",
			err: &net.OpError{
				Op:  "dial",
				Err: errors.New("connection refused"),
			},
			isNetwork: true,
		},
		{
			name:      "should not detect generic errors as network error",
			err:       errors.New("some other error"),
			isNetwork: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNetworkError(tc.err)
			assert.Equal(t, tc.isNetwork, result)
		})
	}
}
