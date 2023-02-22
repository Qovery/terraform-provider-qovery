package git_repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

func TestGitRepositoryValidate(t *testing.T) {
    // setup:
    defaultBranchName := "main"
    defaultRootPath := "/"
    testCases := []struct {
        description string
        url string
        branch *string
        rootPath *string
        expectedError error
    } {
        {description: "case 1: url is blank", url: "", branch: &defaultBranchName, rootPath: &defaultRootPath, expectedError: git_repository.ErrInvalidURLParam},
        {description: "case 2: branch is nil", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: nil, rootPath: &defaultRootPath, expectedError: nil},
        {description: "case 3: rootPath is nil", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: &defaultBranchName, rootPath: nil, expectedError: nil},
        {description: "case 4: all fields are set", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: &defaultBranchName, rootPath: &defaultRootPath, expectedError: nil},
        {description: "case 5: url only is set", url: "https://github.com/Qovery/terraform-provider-qovery.git", branch: nil, rootPath: nil, expectedError: nil},
    }
	
	t.Parallel()
	for _, tc := range testCases {
	   t.Run(tc.description, func(t *testing.T) {
	       // execute:
           i := git_repository.GitRepository{
               Url: tc.url,
               Branch: tc.branch,
               RootPath: tc.rootPath,
            }
           
           // verify:
           assert.Equal(t, tc.expectedError, i.Validate())
		})
	}
}

func TestNewGitRepository(t *testing.T) {
    // setup:
    defaultUrl := "https://github.com/Qovery/terraform-provider-qovery.git"
    defaultBranchName := "main"
    defaultRootPath := "/"
    testCases := []struct {
        description string
        params git_repository.NewGitRepositoryParams
        expectedResult *git_repository.GitRepository
        expectedError error
    } {
        {
            description: "case 1: all params blanks",
            params: git_repository.NewGitRepositoryParams{
                Url: "",
                Branch: nil,
                RootPath: nil,
            },
            expectedError: git_repository.ErrInvalidURLParam,
            expectedResult: nil,
        },
        {
            description: "case 2: url is blank",
            params: git_repository.NewGitRepositoryParams{
                Url: "",
                Branch: &defaultBranchName,
                RootPath: &defaultRootPath,
            },
            expectedError: git_repository.ErrInvalidURLParam,
            expectedResult: nil,
        },
        {
            description: "case 3: branch is blank",
            params: git_repository.NewGitRepositoryParams{
                Url: defaultUrl,
                Branch: nil,
                RootPath: &defaultRootPath,
            },
            expectedError: nil,
            expectedResult: &git_repository.GitRepository {
                Url: defaultUrl,
                Branch: nil,
                RootPath: &defaultRootPath,
            },
        },
        {
            description: "case 4: root path is blank",
            params: git_repository.NewGitRepositoryParams{
                Url: defaultUrl,
                Branch: &defaultBranchName,
                RootPath: nil,
            },
            expectedError: nil,
            expectedResult: &git_repository.GitRepository {
                Url: defaultUrl,
                Branch: &defaultBranchName,
                RootPath: nil,
            },
        },
        {
            description: "case 5: all properly set",
            params: git_repository.NewGitRepositoryParams{
                Url: defaultUrl,
                Branch: &defaultBranchName,
                RootPath: &defaultRootPath,
            },
            expectedError: nil,
            expectedResult: &git_repository.GitRepository {
                Url: defaultUrl,
                Branch: &defaultBranchName,
                RootPath: &defaultRootPath,
            },
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