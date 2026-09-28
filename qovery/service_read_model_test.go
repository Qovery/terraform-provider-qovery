//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

func TestStringListFromAPI(t *testing.T) {
	t.Parallel()

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	argumentList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("--flag"), types.StringValue("value")})

	testCases := []struct {
		TestName string
		Prior    types.List
		API      []string
		Expect   types.List
	}{
		{TestName: "empty_api_value_keeps_a_null_prior", Prior: types.ListNull(types.StringType), API: []string{}, Expect: types.ListNull(types.StringType)},
		{TestName: "nil_api_value_keeps_a_null_prior", Prior: types.ListNull(types.StringType), API: nil, Expect: types.ListNull(types.StringType)},
		{TestName: "empty_api_value_keeps_an_empty_prior", Prior: emptyList, API: []string{}, Expect: emptyList},
		{TestName: "api_value_wins_over_a_null_prior", Prior: types.ListNull(types.StringType), API: []string{"--flag", "value"}, Expect: argumentList},
		{TestName: "api_value_wins_over_the_prior", Prior: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old")}), API: []string{"--flag", "value"}, Expect: argumentList},
		{TestName: "empty_api_value_clears_a_non_empty_prior", Prior: argumentList, API: []string{}, Expect: emptyList},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := stringListFromAPI(tc.Prior, tc.API)
			assert.True(t, tc.Expect.Equal(got), "got %s, want %s", got, tc.Expect)
		})
	}
}

func TestEphemeralStorageFromAPI(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.Int64Value(0), ephemeralStorageFromAPI(nil), "no ephemeral storage reads as the 0 default")
	assert.Equal(t, types.Int64Value(0), ephemeralStorageFromAPI(new(int32(0))))
	assert.Equal(t, types.Int64Value(4), ephemeralStorageFromAPI(new(int32(4))))
}

func TestJobRootPathFromAPI(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.StringValue("/"), jobRootPathFromAPI(nil))
	assert.Equal(t, types.StringValue("/"), jobRootPathFromAPI(new("")), "an API \"\" reads as /")
	assert.Equal(t, types.StringValue("/"), jobRootPathFromAPI(new("/")))
	assert.Equal(t, types.StringValue("/app"), jobRootPathFromAPI(new("/app")))
}

func TestFromCustomDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName                  string
		API                       qovery.CustomDomain
		ExpectGenerateCertificate bool
		ExpectUseCdn              bool
	}{
		{TestName: "api_values_are_reported", API: qovery.CustomDomain{Id: "id", Domain: "app.example.com", GenerateCertificate: true, UseCdn: new(true)}, ExpectGenerateCertificate: true, ExpectUseCdn: true},
		{TestName: "false_values_are_reported", API: qovery.CustomDomain{Id: "id", Domain: "app.example.com", GenerateCertificate: false, UseCdn: new(false)}},
		{TestName: "missing_use_cdn_reads_as_false", API: qovery.CustomDomain{Id: "id", Domain: "app.example.com", GenerateCertificate: true}, ExpectGenerateCertificate: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := fromCustomDomain(&tc.API)
			assert.Equal(t, types.BoolValue(tc.ExpectGenerateCertificate), got.GenerateCertificate)
			assert.Equal(t, types.BoolValue(tc.ExpectUseCdn), got.UseCdn)
			assert.Equal(t, types.StringValue(tc.API.Domain), got.Domain)
		})
	}
}

// TestFromCustomDomainList_IgnoresThePrior covers the gate 0.x had: a flag the prior held as null
// stayed null whatever the API held, so a Console change never showed in the plan.
func TestFromCustomDomainList_IgnoresThePrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := testCustomDomains(testCustomDomain(types.StringValue("id"), types.StringValue("app.example.com"), types.BoolNull(), types.BoolNull()))
	api := []*qovery.CustomDomain{{Id: "id", Domain: "app.example.com", GenerateCertificate: true, UseCdn: new(true)}}

	got := fromCustomDomainList(prior, api).toTerraformSet(ctx)
	require.Len(t, got.Elements(), 1)
	attributes := got.Elements()[0].(types.Object).Attributes()
	assert.Equal(t, types.BoolValue(true), attributes["generate_certificate"])
	assert.Equal(t, types.BoolValue(true), attributes["use_cdn"])

	assert.True(t, fromCustomDomainList(types.SetNull(testCustomDomainType), nil).toTerraformSet(ctx).IsNull(), "no domain keeps a null prior")
	assert.Len(t, fromCustomDomainList(testCustomDomains(), nil).toTerraformSet(ctx).Elements(), 0, "no domain keeps an empty prior")
}

func TestHelmValuesOverrideFromDomain_SetShapes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyMap := types.MapValueMust(types.StringType, map[string]attr.Value{})
	nullMap := types.MapNull(types.StringType)

	testCases := []struct {
		TestName        string
		PriorSet        types.Map
		PriorSetString  types.Map
		ExpectSet       types.Map
		ExpectSetString types.Map
	}{
		// 0.x read set with the shape of set_string, so each of these gave an inconsistent result.
		{TestName: "empty_set_with_null_set_string", PriorSet: emptyMap, PriorSetString: nullMap, ExpectSet: emptyMap, ExpectSetString: nullMap},
		{TestName: "null_set_with_empty_set_string", PriorSet: nullMap, PriorSetString: emptyMap, ExpectSet: nullMap, ExpectSetString: emptyMap},
		{TestName: "both_null", PriorSet: nullMap, PriorSetString: nullMap, ExpectSet: nullMap, ExpectSetString: nullMap},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			prior := &HelmValuesOverride{
				HelmValuesOverrideSet:       tc.PriorSet,
				HelmValuesOverrideSetString: tc.PriorSetString,
				HelmValuesOverrideSetJson:   nullMap,
			}
			got := HelmValuesOverrideFromDomainHelmValuesOverride(ctx, helm.ValuesOverride{}, prior)
			assert.True(t, tc.ExpectSet.Equal(got.HelmValuesOverrideSet), "set: got %s, want %s", got.HelmValuesOverrideSet, tc.ExpectSet)
			assert.True(t, tc.ExpectSetString.Equal(got.HelmValuesOverrideSetString), "set_string: got %s, want %s", got.HelmValuesOverrideSetString, tc.ExpectSetString)
		})
	}
}

func TestHelmValuesOverrideFromDomain_GitTokenID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	withToken := func(token *string) helm.ValuesOverride {
		return helm.ValuesOverride{File: &helm.ValuesOverrideFile{GitRepository: &helm.ValuesOverrideGit{
			Url: "https://github.com/Qovery/helm_chart_engine_testing.git", Branch: "main", Paths: []string{"values.yaml"}, GitToken: token,
		}}}
	}

	noToken := HelmValuesOverrideFromDomainHelmValuesOverride(ctx, withToken(nil), nil)
	require.NotNil(t, noToken.HelmValuesOverrideFile)
	require.NotNil(t, noToken.HelmValuesOverrideFile.GitRepository)
	assert.True(t, noToken.HelmValuesOverrideFile.GitRepository.GitTokenId.IsNull(), "no token reads as null, not \"\"")

	token := HelmValuesOverrideFromDomainHelmValuesOverride(ctx, withToken(new("token-id")), nil)
	assert.Equal(t, types.StringValue("token-id"), token.HelmValuesOverrideFile.GitRepository.GitTokenId)
}

func TestFromVariableArray(t *testing.T) {
	t.Parallel()

	variable := func(key, value string, secret bool) attr.Value {
		return types.ObjectValueMust(terraformVariableAttrTypes, map[string]attr.Value{
			"key":       types.StringValue(key),
			"value":     types.StringValue(value),
			"is_secret": types.BoolValue(secret),
		})
	}
	variables := func(elements ...attr.Value) types.Set {
		return types.SetValueMust(types.ObjectType{AttrTypes: terraformVariableAttrTypes}, elements)
	}
	nullVariables := types.SetNull(types.ObjectType{AttrTypes: terraformVariableAttrTypes})

	testCases := []struct {
		TestName string
		Prior    types.Set
		API      []terraformservice.Variable
		Expect   types.Set
	}{
		{
			TestName: "non_secret_value_changed_outside_terraform_is_reported",
			Prior:    variables(variable("region", "eu-west-3", false)),
			API:      []terraformservice.Variable{{Key: "region", Value: "us-east-1"}},
			Expect:   variables(variable("region", "us-east-1", false)),
		},
		{
			TestName: "non_secret_variable_added_outside_terraform_is_reported",
			Prior:    nullVariables,
			API:      []terraformservice.Variable{{Key: "region", Value: "us-east-1"}},
			Expect:   variables(variable("region", "us-east-1", false)),
		},
		{
			TestName: "secret_value_is_kept_from_the_prior",
			Prior:    variables(variable("token", "s3cr3t", true)),
			API:      []terraformservice.Variable{{Key: "token", Value: "SECRET_VALUE_UNCHANGED", Secret: true}},
			Expect:   variables(variable("token", "s3cr3t", true)),
		},
		{
			TestName: "no_api_variable_keeps_a_null_prior",
			Prior:    nullVariables,
			API:      nil,
			Expect:   nullVariables,
		},
		{
			TestName: "no_api_variable_is_reported_over_a_declared_prior",
			Prior:    variables(variable("region", "eu-west-3", false)),
			API:      nil,
			Expect:   variables(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := fromVariableArray(tc.Prior, tc.API)
			assert.True(t, tc.Expect.Equal(got), "got %s, want %s", got, tc.Expect)
		})
	}
}
