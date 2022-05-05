package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Project struct {
	Id                          types.String `tfsdk:"id"`
	OrganizationId              types.String `tfsdk:"organization_id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	BuiltInEnvironmentVariables types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	Secrets                     types.Set    `tfsdk:"secrets"`
}

func (p Project) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.EnvironmentVariables)
}

func (p Project) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.BuiltInEnvironmentVariables)
}

func (p Project) SecretList() SecretList {
	return toSecretList(p.Secrets)
}

func (p Project) toCreateProjectRequest() client.ProjectUpsertParams {
	return client.ProjectUpsertParams{
		ProjectRequest: qovery.ProjectRequest{
			Name:        toString(p.Name),
			Description: toStringPointer(p.Description),
		},
		EnvironmentVariablesDiff: p.EnvironmentVariableList().diff(nil),
		SecretsDiff:              p.SecretList().diff(nil),
	}
}

func (p Project) toUpdateProjectRequest(state Project) client.ProjectUpsertParams {
	return client.ProjectUpsertParams{
		ProjectRequest: qovery.ProjectRequest{
			Name:        toString(p.Name),
			Description: toStringPointer(p.Description),
		},
		EnvironmentVariablesDiff: p.EnvironmentVariableList().diff(state.EnvironmentVariableList()),
		SecretsDiff:              p.SecretList().diff(state.SecretList()),
	}
}

func convertResponseToProject(state Project, res *client.ProjectResponse) Project {
	return Project{
		Id:                          fromString(res.ProjectResponse.Id),
		OrganizationId:              fromString(res.ProjectResponse.Organization.Id),
		Name:                        fromString(res.ProjectResponse.Name),
		Description:                 fromStringPointer(res.ProjectResponse.Description),
		BuiltInEnvironmentVariables: fromEnvironmentVariableList(res.ProjectEnvironmentVariables, qovery.ENVIRONMENTVARIABLESCOPEENUM_BUILT_IN).toTerraformSet(),
		EnvironmentVariables:        fromEnvironmentVariableList(res.ProjectEnvironmentVariables, qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT).toTerraformSet(),
		Secrets:                     fromSecretList(state.SecretList(), res.ProjectSecret, qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT).toTerraformSet(),
	}
}
