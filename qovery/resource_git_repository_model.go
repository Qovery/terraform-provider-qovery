package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

type GitRepository struct {
	Url        types.String `tfsdk:"url"`
	Branch     types.String `tfsdk:"branch"`
	RootPath   types.String `tfsdk:"root_path"`
	GitTokenId types.String `tfsdk:"git_token_id"`
}

func (g GitRepository) toUpsertRequest() git_repository.GitRepository {
	var branch *string = nil
	if !g.Branch.IsNull() {
		v := ToString(g.Branch)
		branch = &v
	}

	var rootPath *string = nil
	if !g.RootPath.IsNull() {
		v := ToString(g.RootPath)
		rootPath = &v
	}

	var gitTokenId *string = nil
	if !g.GitTokenId.IsNull() {
		v := ToString(g.GitTokenId)
		gitTokenId = &v
	}

	return git_repository.GitRepository{
		Url:        ToString(g.Url),
		Branch:     branch,
		RootPath:   rootPath,
		GitTokenId: gitTokenId,
	}
}
