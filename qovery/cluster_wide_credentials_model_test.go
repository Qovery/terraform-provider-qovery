//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWideAccountID      = "33333333-3333-3333-3333-333333333333"
	testOtherWideAccountID = "44444444-4444-4444-4444-444444444444"
	testSlackWebhookURL    = "https://hooks.slack.com/services/T000/B000/secret"
	testSlackWebhookMask   = "https://hooks.slack.com/services/XXXXXXXXXXX"
)

func wideCredentialsObj(accountID string, slackWebhookURL types.String) types.Object {
	return types.ObjectValueMust(wideCredentialsAttrTypes, map[string]attr.Value{
		wideCredentialsKeyCloudAccountID:            types.StringValue(accountID),
		wideCredentialsKeySlackWebhookURL:           slackWebhookURL,
		wideCredentialsKeyRequireWideAccessToDeploy: types.BoolUnknown(),
	})
}

func wideCredentialsTestCluster(credentialsID string, wideCredentials types.Object) Cluster {
	return Cluster{
		OrganizationId:       types.StringValue("org-123"),
		CredentialsId:        types.StringValue(credentialsID),
		Name:                 types.StringValue("test-cluster"),
		CloudProvider:        types.StringValue("SCW"),
		Region:               types.StringValue("fr-par"),
		KubernetesMode:       types.StringValue("MANAGED"),
		State:                types.StringValue("DEPLOYED"),
		InstanceType:         types.StringValue("DEV1-L"),
		MinRunningNodes:      types.Int64Value(3),
		MaxRunningNodes:      types.Int64Value(10),
		DiskSize:             types.Int64Value(50),
		Features:             types.ObjectNull(createFeaturesAttrTypes()),
		LabelsGroupIds:       types.SetNull(types.StringType),
		UsingWideCredentials: wideCredentials,
	}
}

func TestToQoveryWidePermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		obj             types.Object
		isUpdate        bool
		expected        bool
		restrictive     bool
		accountID       *string
		slackWebhookURL *string
	}{
		{
			name: "create without the block omits wide_permissions",
			obj:  types.ObjectNull(wideCredentialsAttrTypes),
		},
		{
			name:        "update without the block turns restrictive permissions off",
			obj:         types.ObjectNull(wideCredentialsAttrTypes),
			isUpdate:    true,
			expected:    true,
			restrictive: false,
		},
		{
			name:            "block with a slack webhook",
			obj:             wideCredentialsObj(testWideAccountID, types.StringValue(testSlackWebhookURL)),
			expected:        true,
			restrictive:     true,
			accountID:       qovery.PtrString(testWideAccountID),
			slackWebhookURL: qovery.PtrString(testSlackWebhookURL),
		},
		{
			name:        "block without a slack webhook omits it, which removes it",
			obj:         wideCredentialsObj(testWideAccountID, types.StringNull()),
			isUpdate:    true,
			expected:    true,
			restrictive: true,
			accountID:   qovery.PtrString(testWideAccountID),
		},
		{
			name: "unknown block omits wide_permissions",
			obj:  types.ObjectUnknown(wideCredentialsAttrTypes),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toQoveryWidePermissions(tc.obj, tc.isUpdate)
			if !tc.expected {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.NotNil(t, got.UsingRestrictivePermissions)
			assert.Equal(t, tc.restrictive, *got.UsingRestrictivePermissions)
			assert.Equal(t, tc.accountID, got.WideScopeCloudAccountId.Get())
			assert.Equal(t, tc.slackWebhookURL, got.SlackWebhookUrl.Get())
		})
	}
}

func TestCluster_toUpsertClusterRequest_WideCredentials(t *testing.T) {
	t.Parallel()

	withWide := wideCredentialsObj(testWideAccountID, types.StringValue(testSlackWebhookURL))
	noWide := types.ObjectNull(wideCredentialsAttrTypes)

	tests := []struct {
		name  string
		plan  Cluster
		state *Cluster
		// expectCloudProviderRequest tells whether the cloud provider info request is sent;
		// expectedWideAccountID is the wide-scope account it carries, nil for none.
		expectCloudProviderRequest bool
		expectedWideAccountID      *string
		expectForceUpdate          bool
	}{
		{
			// The cloud provider info request follows the creation: without the account, q-core
			// would clear the one the creation set.
			name:                       "create sends the wide-scope account in the cloud provider info request",
			plan:                       wideCredentialsTestCluster("cred-1", withWide),
			expectCloudProviderRequest: true,
			expectedWideAccountID:      qovery.PtrString(testWideAccountID),
		},
		{
			name:                       "create without the block sends no wide-scope account",
			plan:                       wideCredentialsTestCluster("cred-1", noWide),
			expectCloudProviderRequest: true,
		},
		{
			name:  "update without change sends no cloud provider info request",
			plan:  wideCredentialsTestCluster("cred-1", withWide),
			state: ptr(wideCredentialsTestCluster("cred-1", withWide)),
		},
		{
			name:                       "update of the credentials keeps the wide-scope account",
			plan:                       wideCredentialsTestCluster("cred-2", withWide),
			state:                      ptr(wideCredentialsTestCluster("cred-1", withWide)),
			expectCloudProviderRequest: true,
			expectedWideAccountID:      qovery.PtrString(testWideAccountID),
		},
		{
			name:                       "update of the wide-scope account does not redeploy",
			plan:                       wideCredentialsTestCluster("cred-1", withWide),
			state:                      ptr(wideCredentialsTestCluster("cred-1", wideCredentialsObj(testOtherWideAccountID, types.StringNull()))),
			expectCloudProviderRequest: true,
			expectedWideAccountID:      qovery.PtrString(testWideAccountID),
		},
		{
			name:  "update of the slack webhook only does not redeploy",
			plan:  wideCredentialsTestCluster("cred-1", withWide),
			state: ptr(wideCredentialsTestCluster("cred-1", wideCredentialsObj(testWideAccountID, types.StringNull()))),
		},
		{
			name:                       "adding the block redeploys",
			plan:                       wideCredentialsTestCluster("cred-1", withWide),
			state:                      ptr(wideCredentialsTestCluster("cred-1", noWide)),
			expectCloudProviderRequest: true,
			expectedWideAccountID:      qovery.PtrString(testWideAccountID),
			expectForceUpdate:          true,
		},
		{
			name:                       "removing the block redeploys and clears the wide-scope account",
			plan:                       wideCredentialsTestCluster("cred-1", noWide),
			state:                      ptr(wideCredentialsTestCluster("cred-1", withWide)),
			expectCloudProviderRequest: true,
			expectForceUpdate:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			params, err := tc.plan.toUpsertClusterRequest(tc.state)
			require.NoError(t, err)
			require.NotNil(t, params)

			assert.Equal(t, tc.expectForceUpdate, params.ForceUpdate)
			assert.Equal(t, toQoveryWidePermissions(tc.plan.UsingWideCredentials, tc.state != nil), params.ClusterRequest.WidePermissions)

			if !tc.expectCloudProviderRequest {
				assert.Nil(t, params.ClusterCloudProviderRequest)
				return
			}
			require.NotNil(t, params.ClusterCloudProviderRequest)
			if tc.expectedWideAccountID == nil {
				assert.Nil(t, params.ClusterCloudProviderRequest.WidePermissionCredentials)
				return
			}
			require.NotNil(t, params.ClusterCloudProviderRequest.WidePermissionCredentials)
			assert.Equal(t, tc.expectedWideAccountID, params.ClusterCloudProviderRequest.WidePermissionCredentials.Id)
		})
	}
}

func TestFromQoveryWidePermissions(t *testing.T) {
	t.Parallel()

	restricted := func(accountID, slackWebhookURL *string) *qovery.WidePermissionsResponseDto {
		dto := &qovery.WidePermissionsResponseDto{
			UsingRestrictivePermissions: qovery.PtrBool(true),
			RequireWideAccessToDeploy:   qovery.PtrBool(true),
		}
		if accountID != nil {
			dto.SetWideScopeCloudAccountId(*accountID)
		}
		if slackWebhookURL != nil {
			dto.SetSlackWebhookUrl(*slackWebhookURL)
		}
		return dto
	}
	cloudProviderInfo := func(accountID string) *qovery.ClusterCloudProviderInfo {
		return &qovery.ClusterCloudProviderInfo{
			WidePermissionCredentials: &qovery.ClusterCloudProviderInfoCredentials{Id: qovery.PtrString(accountID)},
		}
	}
	priorWithSlack := wideCredentialsObj(testWideAccountID, types.StringValue(testSlackWebhookURL))

	tests := []struct {
		name              string
		widePermissions   *qovery.WidePermissionsResponseDto
		cloudProviderInfo *qovery.ClusterCloudProviderInfo
		prior             types.Object
		mode              clusterReadMode
		expectNull        bool
		expectedAccountID types.String
		expectedSlack     types.String
	}{
		{
			name:       "no wide permissions reads as null",
			prior:      priorWithSlack,
			expectNull: true,
		},
		{
			name:            "restrictive permissions off read as null",
			widePermissions: &qovery.WidePermissionsResponseDto{UsingRestrictivePermissions: qovery.PtrBool(false)},
			prior:           priorWithSlack,
			expectNull:      true,
		},
		{
			name:              "the mask keeps the url of the prior value",
			widePermissions:   restricted(qovery.PtrString(testWideAccountID), qovery.PtrString(testSlackWebhookMask)),
			prior:             priorWithSlack,
			expectedAccountID: types.StringValue(testWideAccountID),
			expectedSlack:     types.StringValue(testSlackWebhookURL),
		},
		{
			name:              "the mask without prior value (import) is stored",
			widePermissions:   restricted(qovery.PtrString(testWideAccountID), qovery.PtrString(testSlackWebhookMask)),
			prior:             types.ObjectNull(wideCredentialsAttrTypes),
			expectedAccountID: types.StringValue(testWideAccountID),
			expectedSlack:     types.StringValue(testSlackWebhookMask),
		},
		{
			name:              "a webhook removed outside Terraform reads as null",
			widePermissions:   restricted(qovery.PtrString(testWideAccountID), nil),
			prior:             priorWithSlack,
			expectedAccountID: types.StringValue(testWideAccountID),
			expectedSlack:     types.StringNull(),
		},
		{
			name:              "the data source reports the mask",
			widePermissions:   restricted(qovery.PtrString(testWideAccountID), qovery.PtrString(testSlackWebhookMask)),
			prior:             priorWithSlack,
			mode:              clusterReadModeDataSource,
			expectedAccountID: types.StringValue(testWideAccountID),
			expectedSlack:     types.StringValue(testSlackWebhookMask),
		},
		{
			// The cluster returned by the edit predates the cloud provider info request.
			name:              "the cloud provider info account wins over the cluster one",
			widePermissions:   restricted(qovery.PtrString(testOtherWideAccountID), nil),
			cloudProviderInfo: cloudProviderInfo(testWideAccountID),
			prior:             types.ObjectNull(wideCredentialsAttrTypes),
			expectedAccountID: types.StringValue(testWideAccountID),
			expectedSlack:     types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := fromQoveryWidePermissions(tc.widePermissions, tc.cloudProviderInfo, tc.prior, tc.mode)
			if tc.expectNull {
				assert.True(t, got.IsNull())
				return
			}
			require.False(t, got.IsNull())
			attrs := got.Attributes()
			assert.Equal(t, tc.expectedAccountID, attrs[wideCredentialsKeyCloudAccountID])
			assert.Equal(t, tc.expectedSlack, attrs[wideCredentialsKeySlackWebhookURL])
			assert.Equal(t, types.BoolValue(true), attrs[wideCredentialsKeyRequireWideAccessToDeploy])
		})
	}
}
