//go:build unit && !integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// When a freshly created service cannot be converted from its API response, the
// repositories fall back to an identity-only entity: ID, environment and name, every other
// field zero-valued. The resource layer still writes that entity to state so Terraform
// taints the service instead of orphaning it, so each state converter must survive the
// zero-valued sub-structures (nil pointers included) and keep the identifiers.

// nullPlanFor decodes an all-null plan of the resource's schema into the model the
// resource reads its own plan into. It is the least a plan can carry and, unlike a
// zero-valued model, its collections are typed nulls, the way Terraform sends them.
func nullPlanFor[T any](t *testing.T, r resource.Resource) T {
	t.Helper()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	objectType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok, "a resource schema is always an object")

	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, attributes)}

	var model T
	diags := plan.Get(ctx, &model)
	require.False(t, diags.HasError(), diags)

	return model
}

func TestConvertDomainHelmToHelm_IdentityOnlyEntity(t *testing.T) {
	t.Parallel()

	plan := nullPlanFor[Helm](t, &helmResource{})
	id, environmentID := uuid.New(), uuid.New()
	identityOnly := &helm.Helm{ID: id, EnvironmentID: environmentID, Name: "partially-created"}

	var state Helm
	require.NotPanics(t, func() {
		state = convertDomainHelmToHelm(context.Background(), plan, identityOnly)
	})

	assert.Equal(t, id.String(), state.ID.ValueString())
	assert.Equal(t, environmentID.String(), state.EnvironmentID.ValueString())
	require.NotNil(t, state.ValuesOverride)
	assert.Nil(t, state.ValuesOverride.HelmValuesOverrideFile, "no values file was ever read from the API")
}

func TestConvertDomainContainerToContainer_IdentityOnlyEntity(t *testing.T) {
	t.Parallel()

	plan := nullPlanFor[Container](t, &containerResource{})
	id, environmentID := uuid.New(), uuid.New()
	identityOnly := &container.Container{ID: id, EnvironmentID: environmentID, Name: "partially-created"}

	var state Container
	require.NotPanics(t, func() {
		state = convertDomainContainerToContainer(context.Background(), plan, identityOnly)
	})

	assert.Equal(t, id.String(), state.ID.ValueString())
	assert.Equal(t, environmentID.String(), state.EnvironmentID.ValueString())
}

func TestConvertDomainJobToJob_IdentityOnlyEntity(t *testing.T) {
	t.Parallel()

	plan := nullPlanFor[Job](t, &jobResource{})
	id, environmentID := uuid.New(), uuid.New()
	identityOnly := &job.Job{ID: id, EnvironmentID: environmentID, Name: "partially-created"}

	var state Job
	require.NotPanics(t, func() {
		state = convertDomainJobToJob(context.Background(), plan, identityOnly)
	})

	assert.Equal(t, id.String(), state.ID.ValueString())
	assert.Equal(t, environmentID.String(), state.EnvironmentID.ValueString())
}

func TestConvertDomainTerraformServiceToTerraformService_IdentityOnlyEntity(t *testing.T) {
	t.Parallel()

	plan := nullPlanFor[TerraformService](t, &terraformServiceResource{})
	id, environmentID := uuid.New(), uuid.New()
	identityOnly := &terraformservice.TerraformService{ID: id, EnvironmentID: environmentID, Name: "partially-created"}

	var state TerraformService
	require.NotPanics(t, func() {
		state = convertDomainTerraformServiceToTerraformService(context.Background(), plan, identityOnly)
	})

	assert.Equal(t, id.String(), state.ID.ValueString())
	assert.Equal(t, environmentID.String(), state.EnvironmentID.ValueString())
}

func TestConvertDomainDeploymentStageToDeploymentStage_IdentityOnlyEntity(t *testing.T) {
	t.Parallel()

	id, environmentID := uuid.New(), uuid.New()
	identityOnly := &deploymentstage.DeploymentStage{ID: id, EnvironmentID: environmentID, Name: "partially-created"}

	var state DeploymentStage
	require.NotPanics(t, func() {
		state = convertDomainDeploymentStageToDeploymentStage(identityOnly, types.StringNull())
	})

	assert.Equal(t, id.String(), state.Id.ValueString())
	assert.Equal(t, environmentID.String(), state.EnvironmentId.ValueString())
}
