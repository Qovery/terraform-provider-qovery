//go:build unit && !integration
// +build unit,!integration

package client

import (
	"encoding/json"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreserveUnmanagedClusterFields checks the edit request resends the current value of every
// field the provider does not manage. The KEDA profiles are checked on the JSON sent to the API:
// client-go's ClusterKeda only models `enabled`, and q-core resets an omitted profile to NORMAL.
func TestPreserveUnmanagedClusterFields(t *testing.T) {
	t.Parallel()

	currentKeda := func(t *testing.T, raw string) *qovery.ClusterKeda {
		var keda qovery.ClusterKeda
		require.NoError(t, json.Unmarshal([]byte(raw), &keda))
		return &keda
	}

	testCases := []struct {
		TestName     string
		RequestKeda  *qovery.ClusterKeda
		CurrentKeda  string
		ExpectedKeda string
	}{
		{
			TestName:     "keda_enabled_resends_current_profiles",
			RequestKeda:  qovery.NewClusterKeda(true),
			CurrentKeda:  `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
			ExpectedKeda: `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
		{
			TestName:     "keda_disabled_resends_current_profiles",
			RequestKeda:  qovery.NewClusterKeda(false),
			CurrentKeda:  `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
			ExpectedKeda: `{"enabled":false,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
		{
			TestName:     "current_keda_without_profiles_sends_enabled_only",
			RequestKeda:  qovery.NewClusterKeda(true),
			CurrentKeda:  `{"enabled":true}`,
			ExpectedKeda: `{"enabled":true}`,
		},
		{
			TestName:     "no_current_keda_sends_enabled_only",
			RequestKeda:  qovery.NewClusterKeda(true),
			ExpectedKeda: `{"enabled":true}`,
		},
		{
			TestName:    "no_request_keda_stays_omitted",
			CurrentKeda: `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			current := &qovery.Cluster{}
			if tc.CurrentKeda != "" {
				current.Keda = currentKeda(t, tc.CurrentKeda)
			}
			request := qovery.ClusterRequest{Keda: tc.RequestKeda}

			preserveUnmanagedClusterFields(&request, current)

			if tc.ExpectedKeda == "" {
				assert.Nil(t, request.Keda)
				return
			}
			body, err := json.Marshal(request.Keda)
			require.NoError(t, err)
			assert.JSONEq(t, tc.ExpectedKeda, string(body))
		})
	}

	t.Run("disk_size_and_metrics_parameters", func(t *testing.T) {
		t.Parallel()

		currentDiskSize := int32(50)
		currentMetrics := &qovery.MetricsParameters{Enabled: qovery.PtrBool(true)}
		current := &qovery.Cluster{DiskSize: &currentDiskSize, MetricsParameters: currentMetrics}

		omitted := qovery.ClusterRequest{}
		preserveUnmanagedClusterFields(&omitted, current)
		assert.Equal(t, &currentDiskSize, omitted.DiskSize)
		assert.Equal(t, currentMetrics, omitted.MetricsParameters)

		plannedDiskSize := int32(100)
		plannedMetrics := &qovery.MetricsParameters{Enabled: qovery.PtrBool(false)}
		planned := qovery.ClusterRequest{DiskSize: &plannedDiskSize, MetricsParameters: plannedMetrics}
		preserveUnmanagedClusterFields(&planned, current)
		assert.Equal(t, &plannedDiskSize, planned.DiskSize)
		assert.Equal(t, plannedMetrics, planned.MetricsParameters)
	})
}
