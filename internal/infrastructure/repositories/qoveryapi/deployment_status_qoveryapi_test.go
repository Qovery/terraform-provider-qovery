//go:build unit && !integration

package qoveryapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

// failingRoundTripper simulates a transport-level failure (context canceled, network
// down): the generated qovery client then returns a nil *http.Response with the error.
type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("simulated transport failure")
}

func newFailingTransportDeploymentStatusAPI() deploymentStatusQoveryAPI {
	cfg := qovery.NewConfiguration()
	cfg.HTTPClient = &http.Client{Transport: failingRoundTripper{}}
	return deploymentStatusQoveryAPI{client: qovery.NewAPIClient(cfg)}
}

const (
	// Short enough that a regression fails in milliseconds rather than hanging the suite,
	// long enough that several polls fit inside the timeout.
	testWaitTimeout      = 200 * time.Millisecond
	testWaitPollInterval = 10 * time.Millisecond
	testWaitSubject      = "environment 3f1c2e9a-0000-4000-8000-000000000000 to reach state RUNNING"
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
		checkErr := errors.New("environment deployment failed")
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

		// Regression for QOV-2299: the loop used to return nil here, so a deployment that
		// never converged within the timeout was recorded as deployed and apply succeeded.
		assert.ErrorIs(t, err, newdeployment.ErrWaitTimeout)
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
			assert.NotErrorIs(t, err, newdeployment.ErrWaitTimeout)
		case <-time.After(2 * time.Second):
			t.Fatal("wait() did not return after context cancellation; it is blocking on the poll ticker instead " +
				"of observing ctx.Done(), which is what leaves terraform apply hanging (and the state lock held) " +
				"when a deployment is interrupted")
		}
	})
}

func TestNewEnvironmentWaitForExpectedDesiredState_TransportErrorDoesNotPanic(t *testing.T) {
	t.Parallel()

	d := newFailingTransportDeploymentStatusAPI()
	// DELETED is the state whose error path inspects response.StatusCode (looking for
	// a 404); a nil response must not panic there.
	f := d.newEnvironmentWaitForExpectedDesiredState(uuid.New(), newdeployment.DELETED)

	ok, err := f(context.Background())
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestDeploymentStatusQoveryAPI_CheckEnvironmentExists_TransportErrorDoesNotPanic(t *testing.T) {
	t.Parallel()

	d := newFailingTransportDeploymentStatusAPI()

	err, statusCode := d.CheckEnvironmentExists(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, 0, statusCode, "no HTTP response was received, so there is no status code to report")
}
