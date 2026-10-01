//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoredDescriptionFromAPI(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.StringValue(""), storedDescriptionFromAPI(nil), "a null from the API reads as the default the plan holds")
	assert.Equal(t, types.StringValue(""), storedDescriptionFromAPI(new("")))
	assert.Equal(t, types.StringValue("set from the Console"), storedDescriptionFromAPI(new("set from the Console")))
}

func TestOptionalStringFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Prior    types.String
		API      *string
		Expect   types.String
	}{
		{TestName: "api_value_wins_over_a_null_prior", Prior: types.StringNull(), API: new("console"), Expect: types.StringValue("console")},
		{TestName: "api_value_wins_over_the_prior", Prior: types.StringValue("old"), API: new("console"), Expect: types.StringValue("console")},
		{TestName: "nil_api_value_clears_the_prior", Prior: types.StringValue("old"), API: nil, Expect: types.StringNull()},
		{TestName: "empty_api_value_keeps_a_null_prior", Prior: types.StringNull(), API: new(""), Expect: types.StringNull()},
		{TestName: "empty_api_value_keeps_an_empty_prior", Prior: types.StringValue(""), API: new(""), Expect: types.StringValue("")},
		{TestName: "empty_api_value_clears_a_non_empty_prior", Prior: types.StringValue("old"), API: new(""), Expect: types.StringValue("")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expect, optionalStringFromAPI(tc.Prior, tc.API))
		})
	}
}

func TestUpgradeOptionalDescriptionFrom0x(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.StringNull(), upgradeOptionalDescriptionFrom0x(types.StringValue("")), "the empty description 0.x stored becomes null")
	assert.Equal(t, types.StringNull(), upgradeOptionalDescriptionFrom0x(types.StringNull()))
	assert.Equal(t, types.StringValue("kept"), upgradeOptionalDescriptionFrom0x(types.StringValue("kept")))
}

// upgradeStateFrom0x runs the schema version 0 upgrader of r on the 0.x state prior, and reads the
// upgraded state into out.
func upgradeStateFrom0x(t *testing.T, r interface {
	resource.Resource
	resource.ResourceWithUpgradeState
}, prior any, out any) {
	t.Helper()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	current := schemaResp.Schema

	upgrader, ok := r.UpgradeState(ctx)[0]
	require.True(t, ok, "no upgrader from version 0")
	priorState := tfsdk.State{Schema: *upgrader.PriorSchema, Raw: tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), nil)}
	require.False(t, priorState.Set(ctx, prior).HasError())

	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: current, Raw: tftypes.NewValue(current.Type().TerraformType(ctx), nil)}}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.False(t, resp.State.Get(ctx, out).HasError())
}

func TestOrganizationUpgradeStateFrom0x(t *testing.T) {
	t.Parallel()

	prior := Organization{Id: types.StringValue("id"), Name: types.StringValue("organization"), Plan: types.StringValue("ENTERPRISE"), Description: types.StringValue("")}
	var got Organization
	upgradeStateFrom0x(t, organizationResource{}, prior, &got)

	assert.Equal(t, types.StringNull(), got.Description, "the empty description 0.x stored becomes null")
	assert.Equal(t, prior.Name, got.Name)
	assert.Equal(t, prior.Plan, got.Plan)
}

func TestGitTokenUpgradeStateFrom0x(t *testing.T) {
	t.Parallel()

	prior := GitToken{
		ID:                 types.StringValue("id"),
		OrganizationId:     types.StringValue("organization-id"),
		Name:               types.StringValue("token"),
		Description:        types.StringValue(""),
		Type:               types.StringValue("GITHUB"),
		Token:              types.StringValue("secret"),
		BitbucketWorkspace: types.StringNull(),
	}
	var got GitToken
	upgradeStateFrom0x(t, gitTokenResource{}, prior, &got)

	assert.Equal(t, types.StringNull(), got.Description, "the empty description 0.x stored becomes null")
	assert.Equal(t, prior.Token, got.Token)
	assert.Equal(t, prior.Type, got.Type)
}
