package qovery

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Environment struct {
	Id                           types.String `tfsdk:"id"`
	ProjectId                    types.String `tfsdk:"project_id"`
	ClusterId                    types.String `tfsdk:"cluster_id"`
	Name                         types.String `tfsdk:"name"`
	Mode                         types.String `tfsdk:"mode"`
	BuiltInEnvironmentVariables  types.List   `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set    `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set    `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set    `tfsdk:"environment_variable_overrides"`
	Secrets                      types.List   `tfsdk:"secrets"`
	SecretAliases                types.Set    `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set    `tfsdk:"secret_overrides"`
	EnvironmentVariableFiles     types.Set    `tfsdk:"environment_variable_files"`
	SecretFiles                  types.Set    `tfsdk:"secret_files"`
	ExternalSecrets              types.Set    `tfsdk:"external_secrets"`
	ExternalSecretFiles          types.Set    `tfsdk:"external_secret_files"`
}

func (e Environment) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(e.EnvironmentVariables)
}

func (e Environment) EnvironmentVariableAliasesList() EnvironmentVariableList {
	return toEnvironmentVariableList(e.EnvironmentVariableAliases)
}

func (e Environment) EnvironmentVariableOverridesList() EnvironmentVariableList {
	return toEnvironmentVariableList(e.EnvironmentVariableOverrides)
}

func (e Environment) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableListFromTerraformList(e.BuiltInEnvironmentVariables)
}

func (e Environment) SecretList() SecretList {
	return ToSecretList(e.Secrets)
}

func (e Environment) SecretAliasesList() SecretList {
	return ToSecretList(e.SecretAliases)
}

func (e Environment) SecretOverridesList() SecretList {
	return ToSecretList(e.SecretOverrides)
}

func (e Environment) EnvironmentVariableFileList() EnvironmentVariableFileList {
	return toEnvironmentVariableFileList(e.EnvironmentVariableFiles)
}

func (e Environment) SecretFileList() SecretFileList {
	return toSecretFileList(e.SecretFiles)
}

func (e Environment) ExternalSecretList() ExternalSecretList {
	return toExternalSecretList(e.ExternalSecrets)
}

func (e Environment) ExternalSecretFileList() ExternalSecretFileList {
	return toExternalSecretFileList(e.ExternalSecretFiles)
}

// toCreateEnvironmentRequest builds the request from the plan. secretValuesWo holds the
// secrets[].value_wo read from the config.
func (e Environment) toCreateEnvironmentRequest(secretValuesWo writeOnlySecretValues) (*environment.CreateServiceRequest, error) {
	mode, err := environment.NewModeFromString(ToString(e.Mode))
	if err != nil {
		return nil, err
	}

	return &environment.CreateServiceRequest{
		EnvironmentCreateRequest: environment.CreateRepositoryRequest{
			Name:      ToString(e.Name),
			ClusterID: ToStringPointer(e.ClusterId),
			Mode:      mode,
		},
		EnvironmentVariables:         e.EnvironmentVariableList().diffRequest(nil),
		EnvironmentVariableAliases:   e.EnvironmentVariableAliasesList().diffRequest(nil),
		EnvironmentVariableOverrides: e.EnvironmentVariableOverridesList().diffRequest(nil),
		EnvironmentVariableFiles:     e.EnvironmentVariableFileList().diffRequest(nil),
		Secrets:                      e.SecretList().diffRequest(nil, secretValuesWo),
		SecretAliases:                e.SecretAliasesList().diffRequest(nil, nil),
		SecretOverrides:              e.SecretOverridesList().diffRequest(nil, nil),
		SecretFiles:                  e.SecretFileList().diffRequest(nil),
		ExternalSecrets:              e.ExternalSecretList().diffRequest(nil),
		ExternalSecretFiles:          e.ExternalSecretFileList().diffRequest(nil),
	}, nil
}

// toUpdateEnvironmentRequest builds the request from the plan. secretValuesWo holds the
// secrets[].value_wo read from the config.
func (e Environment) toUpdateEnvironmentRequest(state Environment, secretValuesWo writeOnlySecretValues) (*environment.UpdateServiceRequest, error) {
	var mode *environment.Mode
	if !e.Mode.IsNull() {
		m, err := environment.NewModeFromString(ToString(e.Mode))
		if err != nil {
			return nil, err
		}
		mode = m
	}

	return &environment.UpdateServiceRequest{
		EnvironmentUpdateRequest: environment.UpdateRepositoryRequest{
			Name: ToStringPointer(e.Name),
			Mode: mode,
		},
		EnvironmentVariables:         e.EnvironmentVariableList().diffRequest(state.EnvironmentVariableList()),
		EnvironmentVariableAliases:   e.EnvironmentVariableAliasesList().diffRequest(state.EnvironmentVariableAliasesList()),
		EnvironmentVariableOverrides: e.EnvironmentVariableOverridesList().diffRequest(state.EnvironmentVariableOverridesList()),
		EnvironmentVariableFiles:     e.EnvironmentVariableFileList().diffRequest(state.EnvironmentVariableFileList()),
		Secrets:                      e.SecretList().diffRequest(state.SecretList(), secretValuesWo),
		SecretAliases:                e.SecretAliasesList().diffRequest(state.SecretAliasesList(), nil),
		SecretOverrides:              e.SecretOverridesList().diffRequest(state.SecretOverridesList(), nil),
		SecretFiles:                  e.SecretFileList().diffRequest(state.SecretFileList()),
		ExternalSecrets:              e.ExternalSecretList().diffRequest(state.ExternalSecretList()),
		ExternalSecretFiles:          e.ExternalSecretFileList().diffRequest(state.ExternalSecretFileList()),
	}, nil
}

func convertDomainEnvironmentToEnvironment(ctx context.Context, state Environment, env *environment.Environment) Environment {
	return Environment{
		Id:                           FromString(env.ID.String()),
		ProjectId:                    FromString(env.ProjectID.String()),
		ClusterId:                    FromString(env.ClusterID.String()),
		Name:                         FromString(env.Name),
		Mode:                         fromClientEnum(env.Mode),
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(env.BuiltInEnvironmentVariables).toTerraformList(ctx),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, env.EnvironmentVariables, variable.ScopeEnvironment, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, env.EnvironmentVariables, variable.ScopeEnvironment, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableOverrides, env.EnvironmentVariables, variable.ScopeEnvironment, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, env.Secrets, variable.ScopeEnvironment, "VALUE").toTerraformList(),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, env.Secrets, variable.ScopeEnvironment, "ALIAS").toTerraformSet(ctx),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, env.Secrets, variable.ScopeEnvironment, "OVERRIDE").toTerraformSet(ctx),
		EnvironmentVariableFiles:     convertDomainVariablesToEnvironmentVariableFileListWithNullableInitialState(ctx, state.EnvironmentVariableFiles, env.EnvironmentVariables, variable.ScopeEnvironment).toTerraformSet(ctx),
		SecretFiles:                  convertDomainSecretsToSecretFileList(state.SecretFiles, env.Secrets, variable.ScopeEnvironment).toTerraformSet(ctx),
		ExternalSecrets:              convertDomainExternalSecretsToExternalSecretList(env.ExternalSecrets, state.ExternalSecrets, variable.ScopeEnvironment).toTerraformSet(ctx),
		ExternalSecretFiles:          convertDomainExternalSecretFilesToExternalSecretFileList(env.ExternalSecretFiles, state.ExternalSecretFiles, variable.ScopeEnvironment).toTerraformSet(ctx),
	}
}

// withoutUndeclaredQoveryAliases drops aliases Qovery created itself, for service outputs, whose
// key prior (plan on apply, state on refresh) does not hold. Kept in state, apply would delete them
// and next deploy recreate them. Data source keeps them: it reports API as is.
func withoutUndeclaredQoveryAliases(prior Environment, env *environment.Environment) *environment.Environment {
	declaredAliases := prior.EnvironmentVariableAliasesList()
	declaredSecretAliases := prior.SecretAliasesList()
	filtered := *env
	filtered.EnvironmentVariables = slices.DeleteFunc(slices.Clone(env.EnvironmentVariables), func(v variable.Variable) bool {
		return v.Type == "ALIAS" && v.IsCreatedByQovery && declaredAliases.find(v.Key) == nil
	})
	filtered.Secrets = slices.DeleteFunc(slices.Clone(env.Secrets), func(s secret.Secret) bool {
		return s.Type == "ALIAS" && s.IsCreatedByQovery && declaredSecretAliases.find(s.Key) == nil
	})
	return &filtered
}
