package qoveryapi

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/qovery/qovery-client-go"
)

// DetectGitProviderFromURL detects the git provider from a git repository URL.
// It returns the corresponding GitProviderEnum or an error if the provider cannot be determined.
func DetectGitProviderFromURL(repositoryURL string) (qovery.GitProviderEnum, error) {
	urlLower := strings.ToLower(repositoryURL)

	switch {
	case strings.Contains(urlLower, "github.com"):
		return qovery.GITPROVIDERENUM_GITHUB, nil
	case strings.Contains(urlLower, "gitlab.com"):
		return qovery.GITPROVIDERENUM_GITLAB, nil
	case strings.Contains(urlLower, "bitbucket.org"):
		return qovery.GITPROVIDERENUM_BITBUCKET, nil
	default:
		return "", fmt.Errorf("unable to detect git provider from URL: %s. Set `provider` in `git_repository` for a self-hosted instance (GitLab self-managed, GitHub Enterprise)", repositoryURL)
	}
}

// ResolveGitProvider returns the configured provider when set, and otherwise detects it from the URL.
// The configured provider is what lets self-hosted instances through: q-core maps GITLAB and GITHUB
// on an unknown host to GitLab self-managed and GitHub Enterprise.
func ResolveGitProvider(repositoryURL string, provider *string) (qovery.GitProviderEnum, error) {
	if provider == nil || *provider == "" {
		return DetectGitProviderFromURL(repositoryURL)
	}
	p, err := qovery.NewGitProviderEnumFromValue(*provider)
	if err != nil {
		return "", err
	}
	// q-core keeps the provider of a github.com, gitlab.com or bitbucket.org host whatever the
	// request says (GitProviderType.create), so a different configured provider would come back
	// changed after apply.
	if hostProvider, ok := gitProviderFromHost(repositoryURL); ok && hostProvider != *p {
		return "", fmt.Errorf("git provider %s conflicts with the repository URL %s, which is a %s repository: set provider = \"%s\" or remove it", *p, repositoryURL, hostProvider, hostProvider)
	}
	return *p, nil
}

// gitProviderFromHost returns the provider q-core infers from the host of a repository URL, as
// GitProviderType.create does: the first provider whose domain the host contains.
func gitProviderFromHost(repositoryURL string) (qovery.GitProviderEnum, bool) {
	u, err := url.Parse(repositoryURL)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "":
		return "", false
	case strings.Contains(host, "github.com"):
		return qovery.GITPROVIDERENUM_GITHUB, true
	case strings.Contains(host, "gitlab.com"):
		return qovery.GITPROVIDERENUM_GITLAB, true
	case strings.Contains(host, "bitbucket.org"):
		return qovery.GITPROVIDERENUM_BITBUCKET, true
	default:
		return "", false
	}
}

// GitProviderFromAPI returns the provider to store for a git repository read from the API: nil
// when the configuration had none and the API holds the provider detected from the URL, so a
// configuration that relies on detection plans no diff. Any other value is returned as-is so a
// provider changed outside Terraform, or an imported self-hosted repository, shows up.
func GitProviderFromAPI(repositoryURL string, api qovery.GitProviderEnum, prior *string) *string {
	if api == "" {
		return prior
	}
	if prior == nil {
		if detected, err := DetectGitProviderFromURL(repositoryURL); err == nil && detected == api {
			return nil
		}
	}
	v := string(api)
	return &v
}

// gitProviderPointer returns the provider read from the API, nil when the API returned none.
func gitProviderPointer(provider qovery.GitProviderEnum) *string {
	if provider == "" {
		return nil
	}
	v := string(provider)
	return &v
}
