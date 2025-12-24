//go:build unit && !integration
// +build unit,!integration

package gittoken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
)

func TestGitTokenType_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenType   gittoken.GitTokenType
		expectError bool
	}{
		{
			name:        "valid GITHUB type",
			tokenType:   gittoken.GITHUB,
			expectError: false,
		},
		{
			name:        "valid GITLAB type",
			tokenType:   gittoken.GITLAB,
			expectError: false,
		},
		{
			name:        "valid BITBUCKET type",
			tokenType:   gittoken.BITBUCKET,
			expectError: false,
		},
		{
			name:        "invalid type",
			tokenType:   gittoken.GitTokenType("INVALID"),
			expectError: true,
		},
		{
			name:        "empty type",
			tokenType:   gittoken.GitTokenType(""),
			expectError: true,
		},
		{
			name:        "lowercase github",
			tokenType:   gittoken.GitTokenType("github"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tokenType.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewGitTokenTypeFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectedResult gittoken.GitTokenType
	}{
		{
			name:           "valid GITHUB",
			input:          "GITHUB",
			expectError:    false,
			expectedResult: gittoken.GITHUB,
		},
		{
			name:           "valid GITLAB",
			input:          "GITLAB",
			expectError:    false,
			expectedResult: gittoken.GITLAB,
		},
		{
			name:           "valid BITBUCKET",
			input:          "BITBUCKET",
			expectError:    false,
			expectedResult: gittoken.BITBUCKET,
		},
		{
			name:        "invalid type",
			input:       "INVALID",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "lowercase github",
			input:       "github",
			expectError: true,
		},
		{
			name:        "mixed case GitHub",
			input:       "GitHub",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gittoken.NewGitTokenTypeFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, *result)
			}
		})
	}
}

func TestGitTokenType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tokenType      gittoken.GitTokenType
		expectedString string
	}{
		{
			name:           "GITHUB string",
			tokenType:      gittoken.GITHUB,
			expectedString: "GITHUB",
		},
		{
			name:           "GITLAB string",
			tokenType:      gittoken.GITLAB,
			expectedString: "GITLAB",
		},
		{
			name:           "BITBUCKET string",
			tokenType:      gittoken.BITBUCKET,
			expectedString: "BITBUCKET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tokenType.String()
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestAllowedGitTokenTypeValues(t *testing.T) {
	t.Parallel()

	// Verify all allowed values are present
	assert.Len(t, gittoken.AllowedGitTokenTypeValues, 3)
	assert.Contains(t, gittoken.AllowedGitTokenTypeValues, gittoken.GITHUB)
	assert.Contains(t, gittoken.AllowedGitTokenTypeValues, gittoken.GITLAB)
	assert.Contains(t, gittoken.AllowedGitTokenTypeValues, gittoken.BITBUCKET)
}
