package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Environment struct {
	Id                           types.String `tfsdk:"id"`
	ProjectId                    types.String `tfsdk:"project_id"`
	ClusterId                    types.String `tfsdk:"cluster_id"`
	Name                         types.String `tfsdk:"name"`
	Mode                         types.String `tfsdk:"mode"`
	BuiltInEnvironmentVariables  types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set    `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set    `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set    `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set    `tfsdk:"secrets"`
	SecretAliases                types.Set    `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set    `tfsdk:"secret_overrides"`
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
	return toEnvironmentVariableList(e.BuiltInEnvironmentVariables)
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

func (e Environment) toCreateEnvironmentRequest() (*environment.CreateServiceRequest, error) {
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
		Secrets:                      e.SecretList().diffRequest(nil),
		SecretAliases:                e.SecretAliasesList().diffRequest(nil),
		SecretOverrides:              e.SecretOverridesList().diffRequest(nil),
	}, nil
}

func (e Environment) toUpdateEnvironmentRequest(state Environment) (*environment.UpdateServiceRequest, error) {
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
		Secrets:                      e.SecretList().diffRequest(state.SecretList()),
		SecretAliases:                e.SecretAliasesList().diffRequest(state.SecretAliasesList()),
		SecretOverrides:              e.SecretOverridesList().diffRequest(state.SecretOverridesList()),
	}, nil
}

func convertDomainEnvironmentToEnvironment(ctx context.Context, state Environment, env *environment.Environment) Environment {
	return Environment{
		Id:                           FromString(env.ID.String()),
		ProjectId:                    FromString(env.ProjectID.String()),
		ClusterId:                    FromString(env.ClusterID.String()),
		Name:                         FromString(env.Name),
		Mode:                         fromClientEnum(env.Mode),
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(ctx, env.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformSet(ctx),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, env.EnvironmentVariables, variable.ScopeEnvironment, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, env.EnvironmentVariables, variable.ScopeEnvironment, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableOverrides, env.EnvironmentVariables, variable.ScopeEnvironment, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, env.Secrets, variable.ScopeEnvironment, "VALUE").toTerraformSet(ctx),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, env.Secrets, variable.ScopeEnvironment, "ALIAS").toTerraformSet(ctx),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, env.Secrets, variable.ScopeEnvironment, "OVERRIDE").toTerraformSet(ctx),
	}
}
