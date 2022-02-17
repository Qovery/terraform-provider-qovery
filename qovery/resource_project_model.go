package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Project struct {
	Id                   types.String          `tfsdk:"id"`
	OrganizationId       types.String          `tfsdk:"organization_id"`
	Name                 types.String          `tfsdk:"name"`
	Description          types.String          `tfsdk:"description"`
	EnvironmentVariables []EnvironmentVariable `tfsdk:"environment_variables"`
}

func (p Project) toUpsertProjectRequest() qovery.ProjectRequest {
	return qovery.ProjectRequest{
		Name:        toString(p.Name),
		Description: toStringPointer(p.Description),
	}
}

func convertResponseToProject(project *qovery.ProjectResponse, variables *qovery.EnvironmentVariableResponseList) Project {
	return Project{
		Id:                   fromString(project.Id),
		OrganizationId:       fromString(project.Organization.Id),
		Name:                 fromString(project.Name),
		Description:          fromStringPointer(project.Description),
		EnvironmentVariables: convertResponseToEnvironmentVariables(variables, EnvironmentVariableScopeProject),
	}
}
