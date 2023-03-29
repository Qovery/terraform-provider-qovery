package qovery

import "github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"

type GitRepository struct {
	Url      string  `tfsdk:"url"`
	Branch   *string `tfsdk:"branch"`
	CommitID *string `tfsdk:"commit_id"`
	RootPath *string `tfsdk:"root_path"`
}

func (g GitRepository) toUpsertRequest() git_repository.GitRepository {
	return git_repository.GitRepository{
		Url:      g.Url,
		Branch:   g.Branch,
		CommitID: g.CommitID,
		RootPath: g.RootPath,
	}
}
