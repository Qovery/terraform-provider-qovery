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

func TestWait(t *testing.T) {
	t.Parallel()

	t.Run("returns_immediately_when_ok", func(t *testing.T) {
		t.Parallel()
		timeout := time.Hour
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			return true, nil
		}, &timeout)
		assert.NoError(t, err)
	})

	t.Run("returns_promptly_on_context_cancellation_instead_of_waiting_for_timeout", func(t *testing.T) {
		t.Parallel()
		timeout := time.Hour // deliberately long, to prove cancellation isn't waiting for this
		ctx, cancel := context.WithCancel(context.Background())

		firstCallDone := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			first := true
			done <- wait(ctx, func(ctx context.Context) (bool, error) {
				if first {
					first = false
					close(firstCallDone)
				}
				return false, nil // never satisfied: forces the loop to rely on the ticker
			}, &timeout)
		}()

		// wait() always calls f synchronously once before entering the poll loop;
		// cancel right after that so we're exercising ctx.Done() in the select, not
		// racing the goroutine's startup.
		<-firstCallDone
		cancel()

		select {
		case err := <-done:
			assert.ErrorIs(t, err, context.Canceled, "wait() should return ctx.Err() promptly on cancellation")
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
