//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func environmentWithAliases() *environment.Environment {
	return &environment.Environment{
		EnvironmentVariables: variable.Variables{
			{ID: testVariableID(0), Scope: variable.ScopeEnvironment, Key: "DB_HOST", Value: "QOVERY_OUTPUT_JOB_ZABC1234_DB_HOST", Type: "ALIAS", IsCreatedByQovery: true},
			{ID: testVariableID(1), Scope: variable.ScopeEnvironment, Key: "MY_ALIAS", Value: "QOVERY_OUTPUT_JOB_ZABC1234_DB_HOST", Type: "ALIAS"},
		},
		Secrets: secret.Secrets{
			{ID: testVariableID(2), Scope: variable.ScopeEnvironment, Key: "DB_PASSWORD", Type: "ALIAS", IsCreatedByQovery: true},
			{ID: testVariableID(3), Scope: variable.ScopeEnvironment, Key: "MY_SECRET_ALIAS", Type: "ALIAS"},
		},
	}
}

// priorWithAliases is an environment whose prior (plan or state) holds these alias keys. Nil keys
// leave the attribute null.
func priorWithAliases(ctx context.Context, aliasKeys []string, secretAliasKeys []string) Environment {
	var aliases EnvironmentVariableList
	if aliasKeys != nil {
		aliases = EnvironmentVariableList{}
		for _, key := range aliasKeys {
			aliases = append(aliases, EnvironmentVariable{Id: types.StringValue("id-" + key), Key: types.StringValue(key), Value: types.StringValue("TARGET"), Description: types.StringNull()})
		}
	}
	var secretAliases SecretList
	if secretAliasKeys != nil {
		secretAliases = SecretList{}
		for _, key := range secretAliasKeys {
			secretAliases = append(secretAliases, Secret{Id: types.StringValue("id-" + key), Key: types.StringValue(key), Value: types.StringValue("TARGET"), Description: types.StringNull()})
		}
	}
	return Environment{
		BuiltInEnvironmentVariables:  EnvironmentVariableList(nil).toTerraformList(ctx),
		EnvironmentVariables:         EnvironmentVariableList(nil).toTerraformSet(ctx),
		EnvironmentVariableAliases:   aliases.toTerraformSet(ctx),
		EnvironmentVariableOverrides: EnvironmentVariableList(nil).toTerraformSet(ctx),
		Secrets:                      SecretList(nil).toTerraformList(),
		SecretAliases:                secretAliases.toTerraformSet(ctx),
		SecretOverrides:              SecretList(nil).toTerraformSet(ctx),
		EnvironmentVariableFiles:     EnvironmentVariableFileList(nil).toTerraformSet(ctx),
		SecretFiles:                  SecretFileList(nil).toTerraformSet(ctx),
		ExternalSecrets:              ExternalSecretList(nil).toTerraformSet(ctx),
		ExternalSecretFiles:          ExternalSecretFileList(nil).toTerraformSet(ctx),
	}
}

func aliasKeys(env Environment) ([]string, []string) {
	keys := []string{}
	for _, v := range env.EnvironmentVariableAliasesList() {
		keys = append(keys, v.Key.ValueString())
	}
	secretKeys := []string{}
	for _, s := range env.SecretAliasesList() {
		secretKeys = append(secretKeys, s.Key.ValueString())
	}
	return keys, secretKeys
}

// TestEnvironmentAliasesCreatedByQovery checks that qovery_environment leaves the aliases Qovery
// creates for service outputs out of its state unless the prior holds their key, so an apply never
// deletes them, while the data source still reports them.
func TestEnvironmentAliasesCreatedByQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		priorAliasKeys        []string
		priorSecretAliasKeys  []string
		expectedAliases       []string
		expectedSecretAliases []string
	}{
		{
			name:                  "qovery_aliases_the_prior_lacks_are_left_out",
			priorAliasKeys:        []string{"MY_ALIAS"},
			priorSecretAliasKeys:  []string{"MY_SECRET_ALIAS"},
			expectedAliases:       []string{"MY_ALIAS"},
			expectedSecretAliases: []string{"MY_SECRET_ALIAS"},
		},
		{
			name:                  "qovery_aliases_the_prior_holds_are_kept",
			priorAliasKeys:        []string{"DB_HOST", "MY_ALIAS"},
			priorSecretAliasKeys:  []string{"DB_PASSWORD", "MY_SECRET_ALIAS"},
			expectedAliases:       []string{"DB_HOST", "MY_ALIAS"},
			expectedSecretAliases: []string{"DB_PASSWORD", "MY_SECRET_ALIAS"},
		},
		{
			name:                  "null_prior_keeps_user_aliases_only",
			expectedAliases:       []string{"MY_ALIAS"},
			expectedSecretAliases: []string{"MY_SECRET_ALIAS"},
		},
	}

	for _, tc := range testCases {
		t.Run("resource/"+tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			env := environmentWithAliases()
			prior := priorWithAliases(ctx, tc.priorAliasKeys, tc.priorSecretAliasKeys)

			aliases, secretAliases := aliasKeys(convertDomainEnvironmentToEnvironment(ctx, prior, withoutUndeclaredQoveryAliases(prior, env)))

			assert.ElementsMatch(t, tc.expectedAliases, aliases)
			assert.ElementsMatch(t, tc.expectedSecretAliases, secretAliases)
			assert.Len(t, env.EnvironmentVariables, 2)
			assert.Len(t, env.Secrets, 2)
		})
	}

	t.Run("data_source/reports_every_alias", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		aliases, secretAliases := aliasKeys(convertDomainEnvironmentToEnvironment(ctx, priorWithAliases(ctx, nil, nil), environmentWithAliases()))

		assert.ElementsMatch(t, []string{"DB_HOST", "MY_ALIAS"}, aliases)
		assert.ElementsMatch(t, []string{"DB_PASSWORD", "MY_SECRET_ALIAS"}, secretAliases)
	})
}
