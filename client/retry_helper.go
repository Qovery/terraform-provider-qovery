package client

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// apiCallFunc is a function that makes an API call and may return an error
type apiCallFunc func(ctx context.Context) *apierrors.APIError

// retryAPICall wraps an API call with exponential backoff retry logic for transient errors.
// This ensures that temporary network issues (DNS failures, timeouts, EOF, etc.) don't
// cause operations to fail unnecessarily.
//
// The retry logic uses:
// - Up to 3 retry attempts
// - Exponential backoff (2s, 4s, 8s) with jitter
// - Automatic detection of transient vs permanent errors
//
// This helper should be used for all direct API calls that aren't already wrapped
// in a wait function, particularly status check calls that happen outside polling loops.
func retryAPICall(ctx context.Context, f apiCallFunc) *apierrors.APIError {
	// Convert apiCallFunc to waitFunc
	waitF := func(ctx context.Context) (bool, *apierrors.APIError) {
		err := f(ctx)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// Use existing retry logic
	_, apiErr := retryOnTransientError(ctx, waitF)
	return apiErr
}
