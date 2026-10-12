//go:build unit && !integration

package qoveryapi

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
)

func strPtr(v string) *string { return &v }

func TestResolveGitProvider(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		URL           string
		Provider      *string
		Expected      qovery.GitProviderEnum
		ErrorContains string
	}{
		{TestName: "detects_github", URL: "https://github.com/qovery/app.git", Expected: qovery.GITPROVIDERENUM_GITHUB},
		{TestName: "detects_gitlab", URL: "https://gitlab.com/qovery/app.git", Expected: qovery.GITPROVIDERENUM_GITLAB},
		{TestName: "detects_bitbucket", URL: "https://bitbucket.org/qovery/app.git", Expected: qovery.GITPROVIDERENUM_BITBUCKET},
		{TestName: "empty_provider_detects", URL: "https://github.com/qovery/app.git", Provider: strPtr(""), Expected: qovery.GITPROVIDERENUM_GITHUB},
		{TestName: "self_hosted_without_provider_fails", URL: "https://gitlab.example.com/group/app.git", ErrorContains: "Set `provider`"},
		{TestName: "self_hosted_gitlab", URL: "https://gitlab.example.com/group/app.git", Provider: strPtr("GITLAB"), Expected: qovery.GITPROVIDERENUM_GITLAB},
		{TestName: "github_enterprise", URL: "https://git.example.com/org/app.git", Provider: strPtr("GITHUB"), Expected: qovery.GITPROVIDERENUM_GITHUB},
		{TestName: "provider_conflicting_with_github_host", URL: "https://github.com/qovery/app.git", Provider: strPtr("GITLAB"), ErrorContains: "conflicts with the repository URL"},
		{TestName: "provider_conflicting_with_gitlab_host", URL: "https://gitlab.com/qovery/app.git", Provider: strPtr("GITHUB"), ErrorContains: "set provider = \"GITLAB\""},
		{TestName: "provider_conflicting_with_bitbucket_host", URL: "https://bitbucket.org/qovery/app.git", Provider: strPtr("GITLAB"), ErrorContains: "conflicts with the repository URL"},
		{TestName: "provider_matching_public_host", URL: "https://github.com/qovery/app.git", Provider: strPtr("GITHUB"), Expected: qovery.GITPROVIDERENUM_GITHUB},
		{TestName: "public_domain_in_path_only_is_self_hosted", URL: "https://git.example.com/mirrors/github.com/app.git", Provider: strPtr("GITLAB"), Expected: qovery.GITPROVIDERENUM_GITLAB},
		{TestName: "public_domain_subdomain_counts_as_public", URL: "https://www.github.com/qovery/app.git", Provider: strPtr("GITLAB"), ErrorContains: "conflicts with the repository URL"},
		{TestName: "invalid_provider", URL: "https://git.example.com/org/app.git", Provider: strPtr("GITEA"), ErrorContains: "GITEA"},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveGitProvider(tc.URL, tc.Provider)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, got)
		})
	}
}

func TestGitProviderFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		URL      string
		API      qovery.GitProviderEnum
		Prior    *string
		Expected *string
	}{
		{TestName: "detected_and_unset_stays_null", URL: "https://github.com/qovery/app.git", API: qovery.GITPROVIDERENUM_GITHUB},
		{TestName: "detected_and_set_keeps_value", URL: "https://github.com/qovery/app.git", API: qovery.GITPROVIDERENUM_GITHUB, Prior: strPtr("GITHUB"), Expected: strPtr("GITHUB")},
		{TestName: "self_hosted_import_reads_api", URL: "https://gitlab.example.com/group/app.git", API: qovery.GITPROVIDERENUM_GITLAB, Expected: strPtr("GITLAB")},
		{TestName: "changed_outside_terraform_shows", URL: "https://github.com/qovery/app.git", API: qovery.GITPROVIDERENUM_GITLAB, Expected: strPtr("GITLAB")},
		{TestName: "set_value_drift_shows", URL: "https://gitlab.example.com/group/app.git", API: qovery.GITPROVIDERENUM_GITHUB, Prior: strPtr("GITLAB"), Expected: strPtr("GITHUB")},
		{TestName: "empty_api_keeps_prior", URL: "https://gitlab.example.com/group/app.git", Prior: strPtr("GITLAB"), Expected: strPtr("GITLAB")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.Expected, GitProviderFromAPI(tc.URL, tc.API, tc.Prior))
		})
	}
}
