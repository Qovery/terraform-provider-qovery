package test_helper

import ("github.com/qovery/terraform-provider-qovery/internal/domain/git_repository")

var (
    DefaultUrl = "https://github.com/Qovery/terraform-provider-qovery.git"
    DefaultBranchName = "main"
	DefaultRootPath = "/"
    
    /// Exposed to tests needing to get such object without having to know internal sauce magic
    DefaultValidNewGitRepositoryParams = git_repository.NewGitRepositoryParams{
        Url:      DefaultUrl,
        Branch:   &DefaultBranchName,
        RootPath: &DefaultRootPath,
    }
    DefaultValidGitRepository = git_repository.GitRepository{
        Url: DefaultValidNewGitRepositoryParams.Url,
        Branch: DefaultValidNewGitRepositoryParams.Branch,
        RootPath: DefaultValidNewGitRepositoryParams.RootPath,
    }
    /// Exposed to tests needing to get such object without having to know internal sauce magic
    DefaultInvalidNewGitRepositoryParams = git_repository.NewGitRepositoryParams{
        Url:      "",
        Branch:   nil,
        RootPath: nil,
    }
    DefaultInvalidGitRepository = git_repository.GitRepository{
        Url: DefaultInvalidNewGitRepositoryParams.Url,
        Branch: DefaultInvalidNewGitRepositoryParams.Branch,
        RootPath: DefaultInvalidNewGitRepositoryParams.RootPath,
    }
    DefaultInvalidNewGitRepositoryParamsError = git_repository.ErrInvalidURLParam
)