package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
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

func (p Project) toCreateServiceRequest() project.UpsertServiceRequest {
	return project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToStringPointer(p.Description),
		},
		EnvironmentVariables: p.EnvironmentVariableList().diffRequest(nil),
		Secrets:              p.SecretList().diffRequest(nil),
	}
}

func (p Project) toUpdateServiceRequest(state Project) project.UpsertServiceRequest {
	return project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToStringPointer(p.Description),
		},
		EnvironmentVariables: p.EnvironmentVariableList().diffRequest(state.EnvironmentVariableList()),
		Secrets:              p.SecretList().diffRequest(state.SecretList()),
	}
}

func convertDomainProjectToProject(state Project, res *project.Project) Project {
	return Project{
		Id:                          FromString(res.ID.String()),
		OrganizationId:              FromString(res.OrganizationID.String()),
		Name:                        FromString(res.Name),
		Description:                 FromStringPointer(res.Description),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(res.EnvironmentVariables, variable.ScopeProject).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(res.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), res.Secrets, variable.ScopeProject).toTerraformSet(),
	}
}
