//go:build unit && !integration

package argoCdDestinationClusterMapping_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

func TestUpsertRequest_Validate(t *testing.T) {
	t.Parallel()

	validRequest := func() argoCdDestinationClusterMapping.UpsertRequest {
		return argoCdDestinationClusterMapping.UpsertRequest{
			AgentClusterId:   gofakeit.UUID(),
			ArgocdClusterUrl: gofakeit.URL(),
			ClusterId:        gofakeit.UUID(),
		}
	}

	testCases := []struct {
		TestName    string
		Mutate      func(*argoCdDestinationClusterMapping.UpsertRequest)
		ExpectError bool
	}{
		{
			TestName:    "fail_with_missing_agent_cluster_id",
			Mutate:      func(r *argoCdDestinationClusterMapping.UpsertRequest) { r.AgentClusterId = "" },
			ExpectError: true,
		},
		{
			TestName:    "fail_with_missing_argocd_cluster_url",
			Mutate:      func(r *argoCdDestinationClusterMapping.UpsertRequest) { r.ArgocdClusterUrl = "" },
			ExpectError: true,
		},
		{
			TestName:    "fail_with_missing_cluster_id",
			Mutate:      func(r *argoCdDestinationClusterMapping.UpsertRequest) { r.ClusterId = "" },
			ExpectError: true,
		},
		{
			TestName: "success",
			Mutate:   func(r *argoCdDestinationClusterMapping.UpsertRequest) {},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req := validRequest()
			tc.Mutate(&req)

			err := req.Validate()
			if tc.ExpectError {
				assert.ErrorContains(t, err, argoCdDestinationClusterMapping.ErrInvalidUpsertRequest.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
