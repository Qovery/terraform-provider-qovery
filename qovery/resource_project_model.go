package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client"
)

type Project struct {
	Id                   types.String          `tfsdk:"id"`
	OrganizationId       types.String          `tfsdk:"organization_id"`
	Name                 types.String          `tfsdk:"name"`
	Description          types.String          `tfsdk:"description"`
	EnvironmentVariables []EnvironmentVariable `tfsdk:"environment_variables"`
}

func (p Project) toCreateProjectRequest() client.ProjectUpsertParams {
	return client.ProjectUpsertParams{
		ProjectRequest: qovery.ProjectRequest{
			Name:        toString(p.Name),
			Description: toStringPointer(p.Description),
		},
		EnvironmentVariablesDiff: diffEnvironmentVariables([]EnvironmentVariable{}, p.EnvironmentVariables),
	}
}

func (p Project) toUpdateProjectRequest(state Project) client.ProjectUpsertParams {
	return client.ProjectUpsertParams{
		ProjectRequest: qovery.ProjectRequest{
			Name:        toString(p.Name),
			Description: toStringPointer(p.Description),
		},
		EnvironmentVariablesDiff: diffEnvironmentVariables(state.EnvironmentVariables, p.EnvironmentVariables),
	}
}

func convertResponseToProject(res *client.ProjectResponse) Project {
	return Project{
		Id:                   fromString(res.ProjectResponse.Id),
		OrganizationId:       fromString(res.ProjectResponse.Organization.Id),
		Name:                 fromString(res.ProjectResponse.Name),
		Description:          fromStringPointer(res.ProjectResponse.Description),
		EnvironmentVariables: convertResponseToEnvironmentVariables(res.ProjectEnvironmentVariables),
	}
}
