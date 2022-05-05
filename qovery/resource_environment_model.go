package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
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

func (e Environment) toCreateEnvironmentRequest() (*client.EnvironmentCreateParams, error) {
	mode, err := qovery.NewEnvironmentModeEnumFromValue(toString(e.Mode))
	if err != nil {
		return nil, err
	}

	return &client.EnvironmentCreateParams{
		EnvironmentRequest: qovery.EnvironmentRequest{
			Name:    toString(e.Name),
			Cluster: toStringPointer(e.ClusterId),
			Mode:    mode,
		},
		EnvironmentVariablesDiff: e.EnvironmentVariableList().diff(nil),
		SecretsDiff:              e.SecretList().diff(nil),
	}, nil
}

func (e Environment) toUpdateEnvironmentRequest(state Environment) client.EnvironmentUpdateParams {
	return client.EnvironmentUpdateParams{
		EnvironmentEditRequest: qovery.EnvironmentEditRequest{
			Name: toStringPointer(e.Name),
		},
		EnvironmentVariablesDiff: e.EnvironmentVariableList().diff(state.EnvironmentVariableList()),
		SecretsDiff:              e.SecretList().diff(state.SecretList()),
	}
}

func convertResponseToEnvironment(state Environment, res *client.EnvironmentResponse) Environment {
	return Environment{
		Id:                          fromString(res.EnvironmentResponse.Id),
		ProjectId:                   fromString(res.EnvironmentResponse.Project.Id),
		ClusterId:                   fromString(res.EnvironmentResponse.ClusterId),
		Name:                        fromString(res.EnvironmentResponse.Name),
		Mode:                        fromClientEnum(res.EnvironmentResponse.Mode),
		BuiltInEnvironmentVariables: fromEnvironmentVariableList(res.EnvironmentEnvironmentVariables, qovery.ENVIRONMENTVARIABLESCOPEENUM_BUILT_IN).toTerraformSet(),
		EnvironmentVariables:        fromEnvironmentVariableList(res.EnvironmentEnvironmentVariables, qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT).toTerraformSet(),
		Secrets:                     fromSecretList(state.SecretList(), res.EnvironmentSecret, qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT).toTerraformSet(),
	}
}
