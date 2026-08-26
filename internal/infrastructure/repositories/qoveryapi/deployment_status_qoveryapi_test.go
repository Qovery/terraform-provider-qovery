//go:build unit && !integration

package qoveryapi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

		calls := 0
		done := make(chan error, 1)
		go func() {
			done <- wait(ctx, func(ctx context.Context) (bool, error) {
				calls++
				return false, nil // never satisfied: forces the loop to rely on the ticker
			}, &timeout)
		}()

		// Let the first synchronous call happen, then cancel.
		time.Sleep(50 * time.Millisecond)
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
