package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Environment struct {
	Id                   types.String          `tfsdk:"id"`
	ProjectId            types.String          `tfsdk:"project_id"`
	ClusterId            types.String          `tfsdk:"cluster_id"`
	Name                 types.String          `tfsdk:"name"`
	Mode                 types.String          `tfsdk:"mode"`
	EnvironmentVariables []EnvironmentVariable `tfsdk:"environment_variables"`
}

func (e Environment) toCreateEnvironmentRequest() client.EnvironmentCreateParams {
	return client.EnvironmentCreateParams{
		EnvironmentRequest: qovery.EnvironmentRequest{
			Name:    toString(e.Name),
			Cluster: toStringPointer(e.ClusterId),
			Mode:    toStringPointer(e.Mode),
		},
		EnvironmentVariablesDiff: diffEnvironmentVariables([]EnvironmentVariable{}, e.EnvironmentVariables),
	}
}

func (e Environment) toUpdateEnvironmentRequest(state Environment) client.EnvironmentUpdateParams {
	return client.EnvironmentUpdateParams{
		EnvironmentEditRequest: qovery.EnvironmentEditRequest{
			Name: toStringPointer(e.Name),
		},
		EnvironmentVariablesDiff: diffEnvironmentVariables(state.EnvironmentVariables, e.EnvironmentVariables),
	}
}

func convertResponseToEnvironment(res *client.EnvironmentResponse) Environment {
	return Environment{
		Id:                   fromString(res.EnvironmentResponse.Id),
		ProjectId:            fromString(res.EnvironmentResponse.Project.Id),
		ClusterId:            fromString(res.EnvironmentResponse.ClusterId),
		Name:                 fromString(res.EnvironmentResponse.Name),
		Mode:                 fromString(res.EnvironmentResponse.Mode),
		EnvironmentVariables: convertResponseToEnvironmentVariables(res.EnvironmentEnvironmentVariables),
	}
}
