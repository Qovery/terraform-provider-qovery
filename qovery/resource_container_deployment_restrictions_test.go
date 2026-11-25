//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
)

// testDeploymentRestrictionSet holds one rule that excludes docs/.
func testDeploymentRestrictionSet(id types.String) types.Set {
	return types.SetValueMust(deploymentRestrictionObjectType, []attr.Value{DeploymentRestriction{
		Id:    id,
		Mode:  types.StringValue("EXCLUDE"),
		Type:  types.StringValue("PATH"),
		Value: types.StringValue("docs/"),
	}.toTerraformObject()})
}

// testDeploymentRestrictionsValue is the configuration of one rule that excludes docs/.
func testDeploymentRestrictionsValue() tftypes.Value {
	setType := testContainerType().AttributeTypes["deployment_restrictions"].(tftypes.Set)
	ruleType := setType.ElementType.(tftypes.Object)
	return tftypes.NewValue(setType, []tftypes.Value{tftypes.NewValue(ruleType, map[string]tftypes.Value{
		"id":    tftypes.NewValue(tftypes.String, nil),
		"mode":  tftypes.NewValue(tftypes.String, "EXCLUDE"),
		"type":  tftypes.NewValue(tftypes.String, "PATH"),
		"value": tftypes.NewValue(tftypes.String, "docs/"),
	})})
}

func validateTestContainerConfig(t *testing.T, config map[string]tftypes.Value) []*tfprotov6.Diagnostic {
	t.Helper()
	config["healthchecks"] = testEmptyHealthchecksValue()
	resp, err := providerserver.NewProtocol6(New("test")())().ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{
		TypeName: "qovery_container",
		Config:   testDynamicValue(t, testContainerSchema(t), testContainerValue(t, config)),
	})
	require.NoError(t, err)
	var errs []*tfprotov6.Diagnostic
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			errs = append(errs, d)
		}
	}
	return errs
}

func TestContainerValidateConfig_DeploymentRestrictions(t *testing.T) {
	t.Parallel()

	t.Run("accepted_on_a_git_repository_source", func(t *testing.T) {
		t.Parallel()
		errs := validateTestContainerConfig(t, testGitContainerConfig(map[string]tftypes.Value{
			"deployment_restrictions": testDeploymentRestrictionsValue(),
		}))
		assert.Empty(t, errs)
	})

	t.Run("refused_on_a_registry_image", func(t *testing.T) {
		t.Parallel()
		config := testImageContainerConfig()
		config["deployment_restrictions"] = testDeploymentRestrictionsValue()
		errs := validateTestContainerConfig(t, config)
		require.Len(t, errs, 1)
		assert.Equal(t, "deployment_restrictions needs source.git_repository", errs[0].Summary)
		assert.True(t, errs[0].Attribute.Equal(tftypes.NewAttributePath().WithAttributeName("deployment_restrictions")), "%s", errs[0].Attribute)
	})
}

func TestContainerToUpsertServiceRequest_DeploymentRestrictions(t *testing.T) {
	t.Parallel()

	ruleID := uuid.NewString()
	newPlan := func(restrictions types.Set) Container {
		plan := nullPlanFor[Container](t, &containerResource{})
		plan.Healthchecks = &HealthChecks{}
		plan.Name = types.StringValue("api")
		plan.Source = &ContainerSource{GitRepository: &ContainerSourceGitRepository{
			URL:            types.StringValue(testGitURL),
			Branch:         types.StringValue("main"),
			RootPath:       types.StringValue("/"),
			DockerfilePath: types.StringValue("Dockerfile"),
		}}
		plan.DeploymentRestrictions = restrictions
		return plan
	}

	t.Run("create_sends_the_planned_rules", func(t *testing.T) {
		t.Parallel()
		request, err := newPlan(testDeploymentRestrictionSet(types.StringUnknown())).toUpsertServiceRequest(nil, nil)
		require.NoError(t, err)
		require.Len(t, request.DeploymentRestrictionsDiff.Create, 1)
		assert.Equal(t, "docs/", request.DeploymentRestrictionsDiff.Create[0].Value)
		assert.Empty(t, request.DeploymentRestrictionsDiff.Delete)
	})

	t.Run("removing_the_block_deletes_the_rules", func(t *testing.T) {
		t.Parallel()
		state := newPlan(testDeploymentRestrictionSet(types.StringValue(ruleID)))
		request, err := newPlan(types.SetNull(deploymentRestrictionObjectType)).toUpsertServiceRequest(&state, nil)
		require.NoError(t, err)
		assert.Empty(t, request.DeploymentRestrictionsDiff.Create)
		assert.Equal(t, []string{ruleID}, request.DeploymentRestrictionsDiff.Delete)
	})

	t.Run("an_invalid_mode_is_an_error", func(t *testing.T) {
		t.Parallel()
		invalid := types.SetValueMust(deploymentRestrictionObjectType, []attr.Value{DeploymentRestriction{
			Id:    types.StringUnknown(),
			Mode:  types.StringValue("ONLY"),
			Type:  types.StringValue("PATH"),
			Value: types.StringValue("docs/"),
		}.toTerraformObject()})
		_, err := newPlan(invalid).toUpsertServiceRequest(nil, nil)
		assert.Error(t, err)
	})
}

func TestConvertDomainContainerToContainer_DeploymentRestrictions(t *testing.T) {
	t.Parallel()

	ruleID := uuid.NewString()
	newContainer := func(rules []deploymentrestriction.ServiceDeploymentRestriction) *container.Container {
		return &container.Container{
			ID:                     uuid.MustParse(testContainerID),
			EnvironmentID:          uuid.MustParse(testEnvironmentID),
			Name:                   "api",
			DeploymentRestrictions: rules,
		}
	}

	t.Run("api_rules_reach_a_state_without_any", func(t *testing.T) {
		t.Parallel()
		prior := nullPlanFor[Container](t, &containerResource{})
		state := convertDomainContainerToContainer(context.Background(), prior, newContainer([]deploymentrestriction.ServiceDeploymentRestriction{
			{Id: &ruleID, Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "docs/"},
		}))
		assert.True(t, testDeploymentRestrictionSet(types.StringValue(ruleID)).Equal(state.DeploymentRestrictions), "%s", state.DeploymentRestrictions)
	})

	t.Run("no_api_rules_keep_an_omitted_block_null", func(t *testing.T) {
		t.Parallel()
		prior := nullPlanFor[Container](t, &containerResource{})
		state := convertDomainContainerToContainer(context.Background(), prior, newContainer(nil))
		assert.True(t, state.DeploymentRestrictions.IsNull())
	})
}
