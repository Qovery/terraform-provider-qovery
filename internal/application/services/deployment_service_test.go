//go:build unit && !integration

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
)

const (
	// Short enough that a regression fails in milliseconds rather than hanging the suite,
	// long enough that several polls fit inside the timeout.
	testWaitTimeout      = 200 * time.Millisecond
	testWaitPollInterval = 10 * time.Millisecond
	testWaitSubject      = "resource 3f1c2e9a-0000-4000-8000-000000000000 to reach state DEPLOYED"
)

func TestWait(t *testing.T) {
	t.Parallel()

	t.Run("returns_immediately_when_ok", func(t *testing.T) {
		t.Parallel()
		calls := 0
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			calls++
			return true, nil
		}, testWaitSubject, testWaitTimeout, testWaitPollInterval)
		assert.NoError(t, err)
		assert.Equal(t, 1, calls, "a satisfied first check must not enter the poll loop")
	})

	t.Run("returns_check_error_without_waiting", func(t *testing.T) {
		t.Parallel()
		checkErr := errors.New("status lookup failed")
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			return false, checkErr
		}, testWaitSubject, testWaitTimeout, testWaitPollInterval)
		assert.ErrorIs(t, err, checkErr)
	})

	t.Run("returns_nil_once_the_check_eventually_succeeds", func(t *testing.T) {
		t.Parallel()
		calls := 0
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			calls++
			return calls >= 3, nil // in progress twice, then converged
		}, testWaitSubject, 10*time.Second, testWaitPollInterval)
		assert.NoError(t, err)
		assert.Equal(t, 3, calls, "wait() must keep polling until the check reports ok")
	})

	t.Run("returns_timeout_error_when_the_check_never_succeeds", func(t *testing.T) {
		t.Parallel()
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			return false, nil // never converges
		}, testWaitSubject, testWaitTimeout, testWaitPollInterval)

		// Regression for QOV-2299: the loop used to return nil here, so a resource that never
		// reached its desired state within the timeout was reported as deployed/stopped/deleted.
		assert.ErrorIs(t, err, deployment.ErrWaitTimeout)
		assert.EqualError(t, err, "deployment wait timed out: waited 200ms for "+testWaitSubject)
	})

	t.Run("returns_promptly_on_context_cancellation_instead_of_waiting_for_timeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		firstCallDone := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			first := true
			// deliberately long timeout and poll interval, to prove cancellation isn't waiting for either
			done <- wait(ctx, func(ctx context.Context) (bool, error) {
				if first {
					first = false
					close(firstCallDone)
				}
				return false, nil // never satisfied: forces the loop to rely on the ticker
			}, testWaitSubject, time.Hour, time.Hour)
		}()

		// wait() always calls f synchronously once before entering the poll loop;
		// cancel right after that so we're exercising ctx.Done() in the select, not
		// racing the goroutine's startup.
		<-firstCallDone
		cancel()

		select {
		case err := <-done:
			assert.ErrorIs(t, err, context.Canceled, "wait() should return ctx.Err() promptly on cancellation")
			assert.NotErrorIs(t, err, deployment.ErrWaitTimeout)
		case <-time.After(2 * time.Second):
			t.Fatal("wait() did not return after context cancellation; it is blocking on the poll ticker instead " +
				"of observing ctx.Done(), which leaves terraform apply hanging (and the state lock held) " +
				"when a deploy/stop/restart is interrupted")
		}
	})
}
