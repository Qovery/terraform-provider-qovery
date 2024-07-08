package qovery

import (
	"context"

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
	EnvironmentVariableAliases  types.Set    `tfsdk:"environment_variable_aliases"`
	Secrets                     types.Set    `tfsdk:"secrets"`
	SecretAliases               types.Set    `tfsdk:"secret_aliases"`
}

func (p Project) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.EnvironmentVariables)
}
func (p Project) EnvironmentVariableAliasesList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.EnvironmentVariableAliases)
}

func (p Project) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.BuiltInEnvironmentVariables)
}

func (p Project) SecretList() SecretList {
	return ToSecretList(p.Secrets)
}
func (p Project) SecretAliasesList() SecretList {
	return ToSecretList(p.SecretAliases)
}

func (p Project) toCreateServiceRequest() project.UpsertServiceRequest {
	return project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToStringPointer(p.Description),
		},
		EnvironmentVariables:       p.EnvironmentVariableList().diffRequest(nil),
		EnvironmentVariableAliases: p.EnvironmentVariableAliasesList().diffRequest(nil),
		Secrets:                    p.SecretList().diffRequest(nil),
		SecretAliases:              p.SecretAliasesList().diffRequest(nil),
	}
}

func (p Project) toUpdateServiceRequest(state Project) project.UpsertServiceRequest {
	return project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToStringPointer(p.Description),
		},
		EnvironmentVariables:       p.EnvironmentVariableList().diffRequest(state.EnvironmentVariableList()),
		EnvironmentVariableAliases: p.EnvironmentVariableAliasesList().diffRequest(state.EnvironmentVariableAliasesList()),
		Secrets:                    p.SecretList().diffRequest(state.SecretList()),
		SecretAliases:              p.SecretAliasesList().diffRequest(state.SecretAliasesList()),
	}
}

func convertDomainProjectToProject(ctx context.Context, state Project, res *project.Project) Project {
	return Project{
		Id:                          FromString(res.ID.String()),
		OrganizationId:              FromString(res.OrganizationID.String()),
		Name:                        FromString(res.Name),
		Description:                 FromStringPointer(res.Description),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(ctx, res.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformSet(ctx),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, res.EnvironmentVariables, variable.ScopeProject, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:  convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, res.EnvironmentVariables, variable.ScopeProject, "ALIAS").toTerraformSet(ctx),
		Secrets:                     convertDomainSecretsToSecretList(state.Secrets, res.Secrets, variable.ScopeProject, "VALUE").toTerraformSet(ctx),
		SecretAliases:               convertDomainSecretsToSecretList(state.SecretAliases, res.Secrets, variable.ScopeProject, "ALIAS").toTerraformSet(ctx),
	}
}
