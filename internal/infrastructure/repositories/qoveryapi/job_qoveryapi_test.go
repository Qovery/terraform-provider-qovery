//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
)

func TestJobBuildSettings(t *testing.T) {
	t.Parallel()

	buildSettings := qovery.NewBuildSettings()
	buildSettings.SetTimeoutMaxSec(5400)

	testCases := []struct {
		TestName string
		Response *qovery.JobResponse
		Expected *qovery.BuildSettings
	}{
		{
			TestName: "nil_response",
			Response: nil,
			Expected: nil,
		},
		{
			TestName: "cron_job",
			Response: &qovery.JobResponse{CronJobResponse: &qovery.CronJobResponse{BuildSettings: buildSettings}},
			Expected: buildSettings,
		},
		{
			TestName: "lifecycle_job",
			Response: &qovery.JobResponse{LifecycleJobResponse: &qovery.LifecycleJobResponse{BuildSettings: buildSettings}},
			Expected: buildSettings,
		},
		{
			TestName: "response_without_build_settings",
			Response: &qovery.JobResponse{CronJobResponse: &qovery.CronJobResponse{}},
			Expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, jobBuildSettings(tc.Response))
		})
	}
}
