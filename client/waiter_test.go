//go:build unit && !integration

package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestWait_ReturnsImmediatelyWhenOk(t *testing.T) {
	t.Parallel()
	apiErr := wait(context.Background(), func(ctx context.Context) (bool, *apierrors.APIError) {
		return true, nil
	})
	assert.Nil(t, apiErr)
}

func TestWait_ReturnsPromptlyOnContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())

	firstCallDone := make(chan struct{})
	done := make(chan *apierrors.APIError, 1)
	go func() {
		first := true
		done <- wait(ctx, func(ctx context.Context) (bool, *apierrors.APIError) {
			if first {
				first = false
				close(firstCallDone)
			}
			return false, nil // never satisfied: forces the loop to rely on the ticker
		})
	}()

	// wait() always calls f synchronously once before entering the poll loop;
	// cancel right after that so we're exercising ctx.Done() in the select, not
	// racing the goroutine's startup.
	<-firstCallDone
	cancel()

	select {
	case apiErr := <-done:
		if assert.NotNil(t, apiErr, "wait() should return an error on cancellation, not nil") {
			assert.ErrorIs(t, apiErr, context.Canceled, "cancellation identity must survive APIError wrapping")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("wait() did not return after context cancellation; it is blocking on the poll ticker instead " +
			"of observing ctx.Done(), which leaves terraform apply hanging (and the state lock held) " +
			"when an application/cluster/database wait is interrupted")
	}
}

func TestWait_ReturnsNilOnceCheckEventuallySucceeds(t *testing.T) {
	t.Parallel()
	calls := 0
	apiErr := waitWithTimeout(context.Background(), func(ctx context.Context) (bool, *apierrors.APIError) {
		calls++
		return calls >= 3, nil // in progress twice, then converged
	}, 10*time.Second, 10*time.Millisecond)
	assert.Nil(t, apiErr)
	assert.Equal(t, 3, calls, "wait() must keep polling until the check reports ok")
}

func TestWait_ReturnsTimeoutErrorWhenCheckNeverSucceeds(t *testing.T) {
	t.Parallel()
	apiErr := waitWithTimeout(context.Background(), func(ctx context.Context) (bool, *apierrors.APIError) {
		return false, nil // never converges
	}, 200*time.Millisecond, 10*time.Millisecond)

	// Guards the same invariant as QOV-2299 in the DDD wait loops: an exhausted timeout is an
	// error, never a silent success that lets apply report a non-converged resource as done.
	if assert.NotNil(t, apiErr, "wait() must fail when the timeout elapses before the check succeeds") {
		assert.ErrorContains(t, apiErr, "operation did not complete within 200ms")
	}
}
