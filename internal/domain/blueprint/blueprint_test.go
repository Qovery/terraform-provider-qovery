//go:build unit && !integration

package blueprint_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

func TestUpsertRequestValidate(t *testing.T) {
	t.Parallel()
	valid := blueprint.UpsertRequest{Name: "my-db", Tag: "aws/postgres/17/1.0.0", IconURI: "app://qovery-console/terraform"}

	testCases := []struct {
		name    string
		mutate  func(r *blueprint.UpsertRequest)
		wantErr error
	}{
		{name: "valid", mutate: func(r *blueprint.UpsertRequest) {}},
		{name: "missing name", mutate: func(r *blueprint.UpsertRequest) { r.Name = "" }, wantErr: blueprint.ErrInvalidUpsertRequest},
		{name: "missing tag", mutate: func(r *blueprint.UpsertRequest) { r.Tag = "" }, wantErr: blueprint.ErrInvalidUpsertRequest},
		{name: "blank variable name", mutate: func(r *blueprint.UpsertRequest) { r.Variables = map[string]string{" ": "x"} }, wantErr: blueprint.ErrBlankVariableName},
		{name: "blank secret variable name", mutate: func(r *blueprint.UpsertRequest) { r.SecretVariables = map[string]string{"": "x"} }, wantErr: blueprint.ErrBlankVariableName},
		{
			name: "variable both secret and non-secret",
			mutate: func(r *blueprint.UpsertRequest) {
				r.Variables = map[string]string{"password": "a"}
				r.SecretVariables = map[string]string{"password": "b"}
			},
			wantErr: blueprint.ErrVariableDeclaredTwice,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			request := valid
			tc.mutate(&request)
			err := request.Validate()
			if tc.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.wantErr.Error())
		})
	}
}

func TestDispatchStatus(t *testing.T) {
	t.Parallel()
	for _, s := range []blueprint.DispatchStatus{blueprint.DispatchStatusDeploying, blueprint.DispatchStatusWaitingRunning, blueprint.DispatchStatusCanceling} {
		assert.True(t, s.IsPending(), s)
		assert.False(t, s.IsSuccess() || s.IsFailure(), s)
	}
	assert.True(t, blueprint.DispatchStatusRunning.IsSuccess())
	for _, s := range []blueprint.DispatchStatus{blueprint.DispatchStatusFailed, blueprint.DispatchStatusInternalError, blueprint.DispatchStatusCanceled} {
		assert.True(t, s.IsFailure(), s)
		assert.False(t, s.IsPending() || s.IsSuccess(), s)
	}
}

func TestCatalogVersion(t *testing.T) {
	t.Parallel()
	version, err := blueprint.ParseCatalogVersion("AWS/postgres/17")
	require.NoError(t, err)
	padded, err := blueprint.ParseCatalogVersion(" AWS/postgres/17 ")
	require.NoError(t, err)
	assert.Equal(t, version, padded)
	assert.Equal(t, blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "17"}, version)
	assert.Equal(t, "AWS/postgres/17", version.String())

	for _, invalid := range []string{"", "AWS/postgres", "AWS/postgres/17/4.1.0", "AWS//17"} {
		_, err := blueprint.ParseCatalogVersion(invalid)
		assert.ErrorIs(t, err, blueprint.ErrInvalidCatalogVersion, invalid)
	}

	fromTag, err := blueprint.CatalogVersionFromTag("AWS/postgres/17/4.1.0")
	require.NoError(t, err)
	assert.Equal(t, version, fromTag)
	_, err = blueprint.CatalogVersionFromTag("AWS/postgres/17")
	assert.ErrorIs(t, err, blueprint.ErrInvalidTag)

	assert.True(t, version.SameService(blueprint.CatalogVersion{Provider: "aws", ServiceFamily: "postgres", ServiceVersion: "18"}))
	assert.False(t, version.SameService(blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "mysql", ServiceVersion: "17"}))
	assert.True(t, version.Equal(blueprint.CatalogVersion{Provider: "aws", ServiceFamily: "Postgres", ServiceVersion: "17"}))
	assert.False(t, version.Equal(blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "18"}))
}

func TestBlueprintLastApplyFailed(t *testing.T) {
	t.Parallel()
	assert.False(t, blueprint.Blueprint{}.LastApplyFailed())
	assert.False(t, blueprint.Blueprint{
		LatestDeployment: &blueprint.Dispatch{Status: blueprint.DispatchStatusRunning},
		ServiceStatus:    &blueprint.ServiceStatus{State: "DEPLOYED"},
	}.LastApplyFailed())
	assert.True(t, blueprint.Blueprint{LatestDeployment: &blueprint.Dispatch{Status: blueprint.DispatchStatusFailed}}.LastApplyFailed())
	assert.True(t, blueprint.Blueprint{LatestDeployment: &blueprint.Dispatch{Status: blueprint.DispatchStatusCanceling}}.LastApplyFailed())
	assert.False(t, blueprint.Blueprint{LatestDeployment: &blueprint.Dispatch{Status: blueprint.DispatchStatusDeploying}}.LastApplyFailed())
	assert.True(t, blueprint.Blueprint{
		LatestDeployment: &blueprint.Dispatch{Status: blueprint.DispatchStatusRunning},
		ServiceStatus:    &blueprint.ServiceStatus{State: "DEPLOYMENT_ERROR"},
	}.LastApplyFailed())
}
