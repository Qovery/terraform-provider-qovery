package git_repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository/test_helper"
)

func TestGitRepositoryValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		url           string
		branch        *string
		rootPath      *string
		expectedError error
	}{
		{description: "case 1: url is blank", url: "", branch: &test_helper.DefaultBranchName, rootPath: &test_helper.DefaultRootPath, expectedError: git_repository.ErrInvalidURLParam},
		{description: "case 2: branch is nil", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: nil, rootPath: &test_helper.DefaultRootPath, expectedError: nil},
		{description: "case 3: rootPath is nil", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: &test_helper.DefaultBranchName, rootPath: nil, expectedError: nil},
		{description: "case 4: all fields are set", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: &test_helper.DefaultBranchName, rootPath: &test_helper.DefaultRootPath, expectedError: nil},
		{description: "case 5: url only is set", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: nil, rootPath: nil, expectedError: nil},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i := git_repository.GitRepository{
				Url:      tc.url,
				Branch:   tc.branch,
				RootPath: tc.rootPath,
			}

			// verify:
			assert.Equal(t, tc.expectedError, i.Validate())
		})
	}
}

func TestNewGitRepository(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         git_repository.NewGitRepositoryParams
		expectedResult *git_repository.GitRepository
		expectedError  error
	}{
		{
			description: "case 1: all params blanks",
			params: git_repository.NewGitRepositoryParams{
				Url:      "",
				Branch:   nil,
				RootPath: nil,
			},
			expectedError:  git_repository.ErrInvalidURLParam,
			expectedResult: nil,
		},
		{
			description: "case 2: url is blank",
			params: git_repository.NewGitRepositoryParams{
				Url:      "",
				Branch:   &test_helper.DefaultBranchName,
				RootPath: &test_helper.DefaultRootPath,
			},
			expectedError:  git_repository.ErrInvalidURLParam,
			expectedResult: nil,
		},
		{
			description: "case 3: branch is blank",
			params: git_repository.NewGitRepositoryParams{
				Url:      test_helper.DefaultUrl,
				Branch:   nil,
				RootPath: &test_helper.DefaultRootPath,
			},
			expectedError: nil,
			expectedResult: &git_repository.GitRepository{
				Url:      test_helper.DefaultUrl,
				Branch:   nil,
				RootPath: &test_helper.DefaultRootPath,
			},
		},
		{
			description: "case 4: root path is blank",
			params: git_repository.NewGitRepositoryParams{
				Url:      test_helper.DefaultUrl,
				Branch:   &test_helper.DefaultBranchName,
				RootPath: nil,
			},
			expectedError: nil,
			expectedResult: &git_repository.GitRepository{
				Url:      test_helper.DefaultUrl,
				Branch:   &test_helper.DefaultBranchName,
				RootPath: nil,
			},
		},
		{
			description: "case 5: all properly set",
			params: git_repository.NewGitRepositoryParams{
				Url:      test_helper.DefaultUrl,
				Branch:   &test_helper.DefaultBranchName,
				RootPath: &test_helper.DefaultRootPath,
			},
			expectedError: nil,
			expectedResult: &git_repository.GitRepository{
				Url:      test_helper.DefaultUrl,
				Branch:   &test_helper.DefaultBranchName,
				RootPath: &test_helper.DefaultRootPath,
			},
		},
		{
			description: "case 6: test default valid new git repository params object (making sure it breaks if not true anymore)",
			params: test_helper.DefaultValidNewGitRepositoryParams,
			expectedError: nil,
			expectedResult: &git_repository.GitRepository{
				Url:      test_helper.DefaultValidNewGitRepositoryParams.Url,
				Branch:   test_helper.DefaultValidNewGitRepositoryParams.Branch,
				RootPath: test_helper.DefaultValidNewGitRepositoryParams.RootPath,
			},
		},
		{
			description: "case 7: test default invalid new git repository params object (making sure it breaks if not true anymore)",
			params: test_helper.DefaultInvalidNewGitRepositoryParams,
			expectedError: git_repository.ErrInvalidURLParam,
			expectedResult: nil,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := git_repository.NewGitRepository(tc.params)

			// verify:
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, nil, err)
			}
			assert.Equal(t, tc.expectedResult, i)
		})
	}
}
