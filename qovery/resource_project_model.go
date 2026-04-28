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
	BuiltInEnvironmentVariables types.List   `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	EnvironmentVariableAliases  types.Set    `tfsdk:"environment_variable_aliases"`
	Secrets                     types.Set    `tfsdk:"secrets"`
	SecretAliases               types.Set    `tfsdk:"secret_aliases"`
	EnvironmentVariableFiles    types.Set    `tfsdk:"environment_variable_files"`
	SecretFiles                 types.Set    `tfsdk:"secret_files"`
}

func (p Project) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.EnvironmentVariables)
}

func (p Project) EnvironmentVariableAliasesList() EnvironmentVariableList {
	return toEnvironmentVariableList(p.EnvironmentVariableAliases)
}

func (p Project) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableListFromTerraformList(p.BuiltInEnvironmentVariables)
}

func (p Project) SecretList() SecretList {
	return ToSecretList(p.Secrets)
}

func (p Project) SecretAliasesList() SecretList {
	return ToSecretList(p.SecretAliases)
}

func (p Project) EnvironmentVariableFileList() EnvironmentVariableFileList {
	return toEnvironmentVariableFileList(p.EnvironmentVariableFiles)
}

func (p Project) SecretFileList() SecretFileList {
	return toSecretFileList(p.SecretFiles)
}

func (p Project) toCreateServiceRequest() project.UpsertServiceRequest {
	return project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToStringPointer(p.Description),
		},
		EnvironmentVariables:       p.EnvironmentVariableList().diffRequest(nil),
		EnvironmentVariableAliases: p.EnvironmentVariableAliasesList().diffRequest(nil),
		EnvironmentVariableFiles:   p.EnvironmentVariableFileList().diffRequest(nil),
		Secrets:                    p.SecretList().diffRequest(nil),
		SecretAliases:              p.SecretAliasesList().diffRequest(nil),
		SecretFiles:                p.SecretFileList().diffRequest(nil),
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
		EnvironmentVariableFiles:   p.EnvironmentVariableFileList().diffRequest(state.EnvironmentVariableFileList()),
		Secrets:                    p.SecretList().diffRequest(state.SecretList()),
		SecretAliases:              p.SecretAliasesList().diffRequest(state.SecretAliasesList()),
		SecretFiles:                p.SecretFileList().diffRequest(state.SecretFileList()),
	}
}

func convertDomainProjectToProject(ctx context.Context, state Project, res *project.Project) Project {
	return Project{
		Id:                          FromString(res.ID.String()),
		OrganizationId:              FromString(res.OrganizationID.String()),
		Name:                        FromString(res.Name),
		Description:                 FromStringPointer(res.Description),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(res.BuiltInEnvironmentVariables).toTerraformList(ctx),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, res.EnvironmentVariables, variable.ScopeProject, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:  convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, res.EnvironmentVariables, variable.ScopeProject, "ALIAS").toTerraformSet(ctx),
		Secrets:                     convertDomainSecretsToSecretList(state.Secrets, res.Secrets, variable.ScopeProject, "VALUE").toTerraformSet(ctx),
		SecretAliases:               convertDomainSecretsToSecretList(state.SecretAliases, res.Secrets, variable.ScopeProject, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableFiles:    convertDomainVariablesToEnvironmentVariableFileListWithNullableInitialState(ctx, state.EnvironmentVariableFiles, res.EnvironmentVariables, variable.ScopeProject).toTerraformSet(ctx),
		SecretFiles:                 convertDomainSecretsToSecretFileList(state.SecretFiles, res.Secrets, variable.ScopeProject).toTerraformSet(ctx),
	}
}
