package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Environment struct {
	Id                   types.String          `tfsdk:"id"`
	ProjectId            types.String          `tfsdk:"project_id"`
	ClusterId            types.String          `tfsdk:"cluster_id"`
	Name                 types.String          `tfsdk:"name"`
	Mode                 types.String          `tfsdk:"mode"`
	EnvironmentVariables []EnvironmentVariable `tfsdk:"environment_variables"`
}

func (e Environment) toCreateEnvironmentRequest() qovery.EnvironmentRequest {
	return qovery.EnvironmentRequest{
		Name:    toString(e.Name),
		Cluster: toStringPointer(e.ClusterId),
		Mode:    toStringPointer(e.Mode),
	}
}

func (e Environment) toUpdateEnvironmentRequest() qovery.EnvironmentEditRequest {
	return qovery.EnvironmentEditRequest{
		Name: toStringPointer(e.Name),
	}
}

func convertResponseToEnvironment(environment *qovery.EnvironmentResponse, variables *qovery.EnvironmentVariableResponseList) Environment {
	return Environment{
		Id:                   fromString(environment.Id),
		ProjectId:            fromString(environment.Project.Id),
		ClusterId:            fromString(environment.ClusterId),
		Name:                 fromString(environment.Name),
		Mode:                 fromString(environment.Mode),
		EnvironmentVariables: convertResponseToEnvironmentVariables(variables, EnvironmentVariableScopeEnvironment),
	}
}
