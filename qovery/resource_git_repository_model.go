package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

type GitRepository struct {
	Url      types.String `tfsdk:"url"`
	Branch   types.String `tfsdk:"branch"`
	RootPath types.String `tfsdk:"root_path"`
}

func (g GitRepository) toUpsertRequest() git_repository.GitRepository {
	var branch *string = nil
	if !g.Branch.IsNull() {
		v := g.Branch.String()
		branch = &v
	}

	var rootPath *string = nil
	if !g.RootPath.IsNull() {
		v := g.RootPath.String()
		rootPath = &v
	}

	return git_repository.GitRepository{
		Url:      g.Url.String(),
		Branch:   branch,
		RootPath: rootPath,
	}
}
