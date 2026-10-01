//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// descriptionAPIEntry is a variable as the API returns it. A nil description means the API
// returns none (the domain layer turns that into "").
type descriptionAPIEntry struct {
	key         string
	description *string
}

// descriptionPriorEntry is a variable as the prior value (plan or state) holds it.
type descriptionPriorEntry struct {
	key         string
	value       types.String
	mountPath   types.String
	description types.String
}

// descriptionResult is what a converter stores for one variable.
type descriptionResult struct {
	value       types.String
	mountPath   types.String
	description types.String
}

// descriptionConverter runs one variable converter. A nil prior is a null prior set. It returns
// nil when the converter stores a null set.
type descriptionConverter func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult

func testVariableID(i int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1))
}

func domainDescription(e descriptionAPIEntry) string {
	if e.description == nil {
		return ""
	}
	return *e.description
}

func clientDescription(e descriptionAPIEntry) qovery.NullableString {
	return *qovery.NewNullableString(e.description)
}

var variableDescriptionConverters = map[string]descriptionConverter{
	"environment_variable_client": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList EnvironmentVariableList
		if prior != nil {
			priorList = EnvironmentVariableList{}
			for _, p := range prior {
				priorList = append(priorList, EnvironmentVariable{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, Description: p.description})
			}
		}
		vars := make([]*qovery.EnvironmentVariable, 0, len(api))
		for i, a := range api {
			vars = append(vars, &qovery.EnvironmentVariable{Id: testVariableID(i).String(), Key: a.key, Value: strPtr("value"), Description: clientDescription(a), Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION, VariableType: qovery.APIVARIABLETYPEENUM_VALUE})
		}
		list := fromEnvironmentVariableListWithNullableInitialState(ctx, priorList.toTerraformSet(ctx), vars, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, description: v.Description}
		}
		return results
	},
	"environment_variable_domain": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList EnvironmentVariableList
		if prior != nil {
			priorList = EnvironmentVariableList{}
			for _, p := range prior {
				priorList = append(priorList, EnvironmentVariable{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, Description: p.description})
			}
		}
		vars := make(variable.Variables, 0, len(api))
		for i, a := range api {
			vars = append(vars, variable.Variable{ID: testVariableID(i), Key: a.key, Value: "value", Description: domainDescription(a), Scope: variable.ScopeContainer, Type: "VALUE"})
		}
		list := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, priorList.toTerraformSet(ctx), vars, variable.ScopeContainer, "VALUE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, description: v.Description}
		}
		return results
	},
	"environment_variable_file_client": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList EnvironmentVariableFileList
		if prior != nil {
			priorList = EnvironmentVariableFileList{}
			for _, p := range prior {
				priorList = append(priorList, EnvironmentVariableFile{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, MountPath: p.mountPath, Description: p.description})
			}
		}
		vars := make([]*qovery.EnvironmentVariable, 0, len(api))
		for i, a := range api {
			vars = append(vars, &qovery.EnvironmentVariable{Id: testVariableID(i).String(), Key: a.key, Value: strPtr("value"), MountPath: *qovery.NewNullableString(strPtr("/etc/" + a.key)), Description: clientDescription(a), Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION, VariableType: qovery.APIVARIABLETYPEENUM_FILE})
		}
		list := fromEnvironmentVariableFileList(ctx, priorList.toTerraformSet(ctx), vars, qovery.APIVARIABLESCOPEENUM_APPLICATION, "FILE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, mountPath: v.MountPath, description: v.Description}
		}
		return results
	},
	"environment_variable_file_domain": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList EnvironmentVariableFileList
		if prior != nil {
			priorList = EnvironmentVariableFileList{}
			for _, p := range prior {
				priorList = append(priorList, EnvironmentVariableFile{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, MountPath: p.mountPath, Description: p.description})
			}
		}
		vars := make(variable.Variables, 0, len(api))
		for i, a := range api {
			vars = append(vars, variable.Variable{ID: testVariableID(i), Key: a.key, Value: "value", MountPath: "/etc/" + a.key, Description: domainDescription(a), Scope: variable.ScopeContainer, Type: "FILE"})
		}
		list := convertDomainVariablesToEnvironmentVariableFileListWithNullableInitialState(ctx, priorList.toTerraformSet(ctx), vars, variable.ScopeContainer)
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, mountPath: v.MountPath, description: v.Description}
		}
		return results
	},
	"secret_client": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList SecretList
		if prior != nil {
			priorList = SecretList{}
			for _, p := range prior {
				priorList = append(priorList, Secret{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, Description: p.description})
			}
		}
		secretType := qovery.APIVARIABLETYPEENUM_VALUE
		secrets := make([]*qovery.Secret, 0, len(api))
		for i, a := range api {
			secrets = append(secrets, &qovery.Secret{Id: testVariableID(i).String(), Key: a.key, Description: clientDescription(a), Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION, VariableType: &secretType})
		}
		list := fromSecretList(priorList.toTerraformSet(ctx), secrets, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, description: v.Description}
		}
		return results
	},
	"secret_domain": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList SecretList
		if prior != nil {
			priorList = SecretList{}
			for _, p := range prior {
				priorList = append(priorList, Secret{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, Description: p.description})
			}
		}
		secrets := make(secret.Secrets, 0, len(api))
		for i, a := range api {
			secrets = append(secrets, secret.Secret{ID: testVariableID(i), Key: a.key, Description: domainDescription(a), Scope: variable.ScopeContainer, Type: "VALUE"})
		}
		list := convertDomainSecretsToSecretList(priorList.toTerraformSet(ctx), secrets, variable.ScopeContainer, "VALUE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, description: v.Description}
		}
		return results
	},
	"secret_file_client": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList SecretFileList
		if prior != nil {
			priorList = SecretFileList{}
			for _, p := range prior {
				priorList = append(priorList, SecretFile{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, MountPath: p.mountPath, Description: p.description})
			}
		}
		secretType := qovery.APIVARIABLETYPEENUM_FILE
		secrets := make([]*qovery.Secret, 0, len(api))
		for i, a := range api {
			secrets = append(secrets, &qovery.Secret{Id: testVariableID(i).String(), Key: a.key, Description: clientDescription(a), Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION, VariableType: &secretType})
		}
		list := fromSecretFileList(priorList.toTerraformSet(ctx), secrets, qovery.APIVARIABLESCOPEENUM_APPLICATION, "FILE")
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, mountPath: v.MountPath, description: v.Description}
		}
		return results
	},
	"secret_file_domain": func(ctx context.Context, prior []descriptionPriorEntry, api []descriptionAPIEntry) map[string]descriptionResult {
		var priorList SecretFileList
		if prior != nil {
			priorList = SecretFileList{}
			for _, p := range prior {
				priorList = append(priorList, SecretFile{Id: types.StringValue("id-" + p.key), Key: types.StringValue(p.key), Value: p.value, MountPath: p.mountPath, Description: p.description})
			}
		}
		// The legacy Secret API returns no mount path.
		secrets := make(secret.Secrets, 0, len(api))
		for i, a := range api {
			secrets = append(secrets, secret.Secret{ID: testVariableID(i), Key: a.key, Description: domainDescription(a), Scope: variable.ScopeContainer, Type: "FILE"})
		}
		list := convertDomainSecretsToSecretFileList(priorList.toTerraformSet(ctx), secrets, variable.ScopeContainer)
		if list == nil {
			return nil
		}
		results := map[string]descriptionResult{}
		for _, v := range list {
			results[v.Key.ValueString()] = descriptionResult{value: v.Value, mountPath: v.MountPath, description: v.Description}
		}
		return results
	},
}

// TestVariableDescriptionsFromAPI covers the description read of environment variables (and
// their aliases and overrides), environment variable files, secrets and secret files, on both
// the application (client) and the domain read paths. The API description always wins; an API
// "" is stored as null unless the prior entry held "". The prior only decides the shape.
func TestVariableDescriptionsFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		prior []descriptionPriorEntry
		api   []descriptionAPIEntry
		want  map[string]types.String
	}{
		{
			// Import and data sources: nothing in the prior, every description comes from the API.
			name:  "prior_null_with_api_values",
			prior: nil,
			api: []descriptionAPIEntry{
				{key: "SET_IN_CONSOLE", description: strPtr("set in the Console")},
				{key: "EMPTY", description: strPtr("")},
				{key: "NONE", description: nil},
			},
			want: map[string]types.String{
				"SET_IN_CONSOLE": types.StringValue("set in the Console"),
				"EMPTY":          types.StringNull(),
				"NONE":           types.StringNull(),
			},
		},
		{
			name:  "prior_null_with_empty_api_value",
			prior: nil,
			api:   nil,
			want:  nil,
		},
		{
			name:  "prior_empty_with_empty_api_value",
			prior: []descriptionPriorEntry{},
			api:   nil,
			want:  map[string]types.String{},
		},
		{
			name: "prior_set",
			prior: []descriptionPriorEntry{
				{key: "SET_IN_CONSOLE", description: types.StringNull()},
				{key: "CHANGED_IN_CONSOLE", description: types.StringValue("declared")},
				{key: "CLEARED_IN_CONSOLE", description: types.StringValue("declared")},
				{key: "DECLARED_EMPTY", description: types.StringValue("")},
				{key: "DECLARED_EMPTY_API_NONE", description: types.StringValue("")},
				{key: "OMITTED", description: types.StringNull()},
			},
			api: []descriptionAPIEntry{
				{key: "SET_IN_CONSOLE", description: strPtr("set in the Console")},
				{key: "CHANGED_IN_CONSOLE", description: strPtr("changed in the Console")},
				{key: "CLEARED_IN_CONSOLE", description: strPtr("")},
				{key: "DECLARED_EMPTY", description: strPtr("")},
				{key: "DECLARED_EMPTY_API_NONE", description: nil},
				{key: "OMITTED", description: strPtr("")},
				{key: "CREATED_IN_CONSOLE", description: strPtr("created in the Console")},
			},
			want: map[string]types.String{
				"SET_IN_CONSOLE":          types.StringValue("set in the Console"),
				"CHANGED_IN_CONSOLE":      types.StringValue("changed in the Console"),
				"CLEARED_IN_CONSOLE":      types.StringValue(""),
				"DECLARED_EMPTY":          types.StringValue(""),
				"DECLARED_EMPTY_API_NONE": types.StringValue(""),
				"OMITTED":                 types.StringNull(),
				"CREATED_IN_CONSOLE":      types.StringValue("created in the Console"),
			},
		},
	}

	for converterName, convert := range variableDescriptionConverters {
		convert := convert
		for _, tc := range testCases {
			tc := tc
			t.Run(converterName+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				results := convert(context.Background(), tc.prior, tc.api)
				if tc.want == nil {
					assert.Nil(t, results)
					return
				}
				require.NotNil(t, results)
				got := make(map[string]types.String, len(results))
				for key, r := range results {
					got[key] = r.description
				}
				assert.Equal(t, tc.want, got)
			})
		}
	}
}

// TestSecretValuesFromPrior covers the secret attributes the API never returns: the value (and,
// on the legacy Secret API, the file mount path) come from the prior entry, and are null when
// there is none (import, data sources).
func TestSecretValuesFromPrior(t *testing.T) {
	t.Parallel()

	prior := []descriptionPriorEntry{
		{key: "DECLARED", value: types.StringValue("s3cr3t"), mountPath: types.StringValue("/etc/declared"), description: types.StringNull()},
	}
	api := []descriptionAPIEntry{{key: "DECLARED"}, {key: "CREATED_IN_CONSOLE"}}

	for _, converterName := range []string{"secret_client", "secret_domain", "secret_file_client", "secret_file_domain"} {
		converterName := converterName
		isFile := converterName == "secret_file_client" || converterName == "secret_file_domain"
		t.Run(converterName, func(t *testing.T) {
			t.Parallel()
			results := variableDescriptionConverters[converterName](context.Background(), prior, api)
			require.Len(t, results, 2)

			assert.Equal(t, types.StringValue("s3cr3t"), results["DECLARED"].value)
			assert.True(t, results["CREATED_IN_CONSOLE"].value.IsNull())
			if isFile {
				assert.Equal(t, types.StringValue("/etc/declared"), results["DECLARED"].mountPath)
				assert.Equal(t, types.StringValue(""), results["CREATED_IN_CONSOLE"].mountPath)
			}
		})
	}
}

// TestVariableUpdateRequestsSendPlannedDescription covers the write side of the description
// contract: an update sends the planned description, and a null one clears it (null on the
// application API, "" on the domain API).
func TestVariableUpdateRequestsSendPlannedDescription(t *testing.T) {
	t.Parallel()

	old := types.StringValue("set in the Console")
	testCases := []struct {
		name    string
		planned types.String
		// wantClient is the description the application API request carries; nil is null.
		wantClient *string
		wantDomain string
	}{
		{name: "planned_value", planned: types.StringValue("declared"), wantClient: strPtr("declared"), wantDomain: "declared"},
		{name: "planned_null_clears", planned: types.StringNull(), wantClient: nil, wantDomain: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clientDescriptions := map[string]qovery.NullableString{
				"environment_variable":      EnvironmentVariable{Description: old}.toUpdateRequest(EnvironmentVariable{Description: tc.planned}).Description,
				"environment_variable_file": EnvironmentVariableFile{Description: old}.toUpdateRequest(EnvironmentVariableFile{Description: tc.planned}).Description,
				"secret":                    Secret{Description: old}.toUpdateRequest(Secret{Description: tc.planned}).Description,
				"secret_file":               SecretFile{Description: old}.toUpdateRequest(SecretFile{Description: tc.planned}).Description,
			}
			for name, description := range clientDescriptions {
				assert.True(t, description.IsSet(), name)
				assert.Equal(t, tc.wantClient, description.Get(), name)
			}

			domainDescriptions := map[string]string{
				"environment_variable":      EnvironmentVariable{Description: old}.toDiffUpdateRequest(EnvironmentVariable{Description: tc.planned}).Description,
				"environment_variable_file": EnvironmentVariableFile{Description: old}.toDiffUpdateRequest(EnvironmentVariableFile{Description: tc.planned}).Description,
				"secret":                    Secret{Description: old}.toDiffUpdateRequest(Secret{Description: tc.planned}).Description,
				"secret_file":               SecretFile{Description: old}.toDiffUpdateRequest(SecretFile{Description: tc.planned}).Description,
			}
			for name, description := range domainDescriptions {
				assert.Equal(t, tc.wantDomain, description, name)
			}
		})
	}
}

// TestSecretListDiff_descriptionOnlyChange covers the application write path: a secret whose only
// change is its description is updated.
func TestSecretListDiff_descriptionOnlyChange(t *testing.T) {
	t.Parallel()

	old := SecretList{{Id: types.StringValue("secret-id"), Key: types.StringValue("KEY"), Value: types.StringValue("s3cr3t"), Description: types.StringValue("set in the Console")}}
	planned := SecretList{{Key: types.StringValue("KEY"), Value: types.StringValue("s3cr3t"), Description: types.StringNull()}}

	diff := planned.diff(old)
	require.Len(t, diff.Update, 1)
	assert.Empty(t, diff.Create)
	assert.Empty(t, diff.Delete)
	assert.Equal(t, "secret-id", diff.Update[0].Id)
	assert.Nil(t, diff.Update[0].Description.Get())
}
