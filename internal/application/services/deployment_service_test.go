//go:build unit && !integration

package services

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
		err := wait(context.Background(), func(ctx context.Context) (bool, error) {
			return true, nil
		})
		assert.NoError(t, err)
	})

	t.Run("returns_promptly_on_context_cancellation_instead_of_waiting_for_timeout", func(t *testing.T) {
		t.Parallel()
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
			})
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
				"of observing ctx.Done(), which leaves terraform apply hanging (and the state lock held) " +
				"when a deploy/stop/restart is interrupted")
		}
	})
}
