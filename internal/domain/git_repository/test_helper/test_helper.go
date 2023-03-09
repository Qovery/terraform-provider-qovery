package test_helper

import (
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

var (
	DefaultUrl        = "https://github.com/Qovery/terraform-provider-qovery.git"
	DefaultBranchName = "main"
	DefaultCommitID   = "42e2e5af9d49de268cd1fda3587788da4ace418a"
	DefaultRootPath   = "/"

	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultValidNewGitRepositoryParams = git_repository.NewGitRepositoryParams{
		Url:      DefaultUrl,
		Branch:   &DefaultBranchName,
		CommitID: &DefaultCommitID,
		RootPath: &DefaultRootPath,
	}
	DefaultValidGitRepository = git_repository.GitRepository{
		Url:      DefaultValidNewGitRepositoryParams.Url,
		Branch:   DefaultValidNewGitRepositoryParams.Branch,
		CommitID: DefaultValidNewGitRepositoryParams.CommitID,
		RootPath: DefaultValidNewGitRepositoryParams.RootPath,
	}
	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultInvalidNewGitRepositoryParams = git_repository.NewGitRepositoryParams{
		Url:      "",
		Branch:   nil,
		CommitID: nil,
		RootPath: nil,
	}
	DefaultInvalidGitRepository = git_repository.GitRepository{
		Url:      DefaultInvalidNewGitRepositoryParams.Url,
		Branch:   DefaultInvalidNewGitRepositoryParams.Branch,
		CommitID: DefaultInvalidNewGitRepositoryParams.CommitID,
		RootPath: DefaultInvalidNewGitRepositoryParams.RootPath,
	}
	DefaultInvalidNewGitRepositoryParamsError = git_repository.ErrInvalidURLParam
)
