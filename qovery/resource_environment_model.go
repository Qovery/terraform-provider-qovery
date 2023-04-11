package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Environment struct {
	Id                          types.String `tfsdk:"id"`
	ProjectId                   types.String `tfsdk:"project_id"`
	ClusterId                   types.String `tfsdk:"cluster_id"`
	Name                        types.String `tfsdk:"name"`
	Mode                        types.String `tfsdk:"mode"`
	BuiltInEnvironmentVariables types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	Secrets                     types.Set    `tfsdk:"secrets"`
}

func (e Environment) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(e.EnvironmentVariables)
}

func (e Environment) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(e.BuiltInEnvironmentVariables)
}

func (e Environment) SecretList() SecretList {
	return toSecretList(e.Secrets)
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
		EnvironmentVariables: e.EnvironmentVariableList().diffRequest(nil),
		Secrets:              e.SecretList().diffRequest(nil),
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
		EnvironmentVariables: e.EnvironmentVariableList().diffRequest(state.EnvironmentVariableList()),
		Secrets:              e.SecretList().diffRequest(state.SecretList()),
	}, nil
}

func convertDomainEnvironmentToEnvironment(state Environment, env *environment.Environment) Environment {
	return Environment{
		Id:                          FromString(env.ID.String()),
		ProjectId:                   FromString(env.ProjectID.String()),
		ClusterId:                   FromString(env.ClusterID.String()),
		Name:                        FromString(env.Name),
		Mode:                        fromClientEnum(env.Mode),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(env.EnvironmentVariables, variable.ScopeEnvironment).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(env.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), env.Secrets, variable.ScopeEnvironment).toTerraformSet(),
	}
}
