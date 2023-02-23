package git_repository

import (
	"github.com/pkg/errors"
)

var (
	// ErrInvalidURLParamdNameParam is returned if the URL param is invalid.
	ErrInvalidURLParam = errors.New("invalid URL param")
)

type GitRepository struct {
	Url      string `validate:"required"`
	Branch   *string
	RootPath *string
}

func (i GitRepository) Validate() error {
	if i.Url == "" {
		return ErrInvalidURLParam
	}

	return nil
}

type NewGitRepositoryParams struct {
	Url      string
	Branch   *string
	RootPath *string
}

func NewGitRepository(params NewGitRepositoryParams) (*GitRepository, error) {
	gitRepository := &GitRepository{
		Url:      params.Url,
		Branch:   params.Branch,
		RootPath: params.RootPath,
	}

	if err := gitRepository.Validate(); err != nil {
		return nil, err
	}

	return gitRepository, nil
}
