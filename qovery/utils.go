package qovery

import (
	"strings"
	"terraform-provider-qovery/qovery/apierror"
	"time"
)

// IsStatusError check if the status state is an Error
func IsStatusError(state string) bool {
	return strings.HasSuffix(state, "_ERROR")
}

func IsFinalState(state string) bool {
	return state != "DEPLOYING" && state != "DELETING" && state != "STOPPING"
}

type WaitCallable func() (bool, *apierror.APIError)

// Wait until timeout (30 minutes)
func Wait(callable WaitCallable) *apierror.APIError {
	return WaitWithTimeout(callable, 30*time.Minute)
}

// WaitWithTimeout wait until timeout
func WaitWithTimeout(callable WaitCallable, timeout time.Duration) *apierror.APIError {
	ticker := time.NewTicker(10 * time.Second)
	mTimeout := time.NewTicker(timeout)

	for {
		select {
		case <-mTimeout.C:
			return nil // silent timeout
		case <-ticker.C:
			res, err := callable()
			if err != nil {
				return err
			}

			if res {
				return nil
			}
		}
	}

}
