//go:build unit && !integration

package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

func TestNewDomainArgoCdDestinationClusterMappingFromResponse(t *testing.T) {
	t.Parallel()

	newResponse := func(agentClusterID, clusterID string) *qovery.ArgoCdDestinationClusterMappingResponse {
		return &qovery.ArgoCdDestinationClusterMappingResponse{
			AgentClusterId:   agentClusterID,
			ArgocdClusterUrl: "https://kubernetes.default.svc",
			ClusterId:        *qovery.NewNullableString(&clusterID),
		}
	}

	testCases := []struct {
		TestName      string
		OrgID         string
		Response      *qovery.ArgoCdDestinationClusterMappingResponse
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "fail_with_invalid_organization_id",
			OrgID:         "not-a-uuid",
			Response:      newResponse(gofakeit.UUID(), gofakeit.UUID()),
			ExpectError:   true,
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidOrganizationIDParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_agent_cluster_id",
			OrgID:         gofakeit.UUID(),
			Response:      newResponse("not-a-uuid", gofakeit.UUID()),
			ExpectError:   true,
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidAgentClusterIDParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_cluster_id",
			OrgID:         gofakeit.UUID(),
			Response:      newResponse(gofakeit.UUID(), "not-a-uuid"),
			ExpectError:   true,
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidClusterIDParam.Error(),
		},
		{
			TestName: "success",
			OrgID:    gofakeit.UUID(),
			Response: newResponse(gofakeit.UUID(), gofakeit.UUID()),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mapping, err := newDomainArgoCdDestinationClusterMappingFromResponse(tc.OrgID, tc.Response)
			if tc.ExpectError {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, mapping)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, mapping)
			assert.Equal(t, tc.OrgID, mapping.OrganizationID.String())
			assert.Equal(t, tc.Response.AgentClusterId, mapping.AgentClusterID.String())
			assert.Equal(t, tc.Response.ArgocdClusterUrl, mapping.ArgocdClusterUrl)
			assert.Equal(t, tc.Response.GetClusterId(), mapping.ClusterID.String())
		})
	}
}
