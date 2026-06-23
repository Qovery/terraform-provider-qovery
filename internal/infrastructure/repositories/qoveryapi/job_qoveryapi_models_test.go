//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

// TestGetAggregateJobResponse_EphemeralStorage guards that the ephemeral storage
// value is carried over from both the cron and lifecycle job responses, and that
// an unset value stays nil (so the platform default is preserved).
func TestGetAggregateJobResponse_EphemeralStorage(t *testing.T) {
	t.Parallel()

	ephemeralStorage := int32(4)

	testCases := []struct {
		TestName string
		Response *qovery.JobResponse
		Expected *int32
	}{
		{
			TestName: "cron_job_with_ephemeral_storage",
			Response: &qovery.JobResponse{
				CronJobResponse: &qovery.CronJobResponse{
					EphemeralStorageInGib: &ephemeralStorage,
				},
			},
			Expected: &ephemeralStorage,
		},
		{
			TestName: "lifecycle_job_with_ephemeral_storage",
			Response: &qovery.JobResponse{
				LifecycleJobResponse: &qovery.LifecycleJobResponse{
					EphemeralStorageInGib: &ephemeralStorage,
				},
			},
			Expected: &ephemeralStorage,
		},
		{
			TestName: "cron_job_without_ephemeral_storage",
			Response: &qovery.JobResponse{
				CronJobResponse: &qovery.CronJobResponse{},
			},
			Expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			aggregate := getAggregateJobResponse(tc.Response)
			assert.Equal(t, tc.Expected, aggregate.EphemeralStorage)
		})
	}
}

// TestNewQoveryJobRequestFromDomain_EphemeralStorage guards the domain→request
// direction: a set value is forwarded and an unset value stays nil (omitempty →
// no payload sent, so the platform default is used).
func TestNewQoveryJobRequestFromDomain_EphemeralStorage(t *testing.T) {
	t.Parallel()

	ephemeralStorage := int32(8)

	testCases := []struct {
		TestName string
		Request  job.UpsertRepositoryRequest
		Expected *int32
	}{
		{
			TestName: "with_ephemeral_storage",
			Request:  job.UpsertRepositoryRequest{Name: "test-job", EphemeralStorage: &ephemeralStorage},
			Expected: &ephemeralStorage,
		},
		{
			TestName: "without_ephemeral_storage",
			Request:  job.UpsertRepositoryRequest{Name: "test-job"},
			Expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req, err := newQoveryJobRequestFromDomain(tc.Request)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, req.EphemeralStorageInGib)
		})
	}
}
