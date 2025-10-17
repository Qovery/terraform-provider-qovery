package qovery

import (
	"fmt"
	"strings"

	"github.com/qovery/qovery-client-go"
)

// detectGitProviderFromURL detects the git provider from a git repository URL.
// It returns the corresponding GitProviderEnum or an error if the provider cannot be determined.
func detectGitProviderFromURL(url string) (qovery.GitProviderEnum, error) {
	urlLower := strings.ToLower(url)

	switch {
	case strings.Contains(urlLower, "github.com"):
		return qovery.GITPROVIDERENUM_GITHUB, nil
	case strings.Contains(urlLower, "gitlab.com"):
		return qovery.GITPROVIDERENUM_GITLAB, nil
	case strings.Contains(urlLower, "bitbucket.org"):
		return qovery.GITPROVIDERENUM_BITBUCKET, nil
	default:
		return "", fmt.Errorf("unable to detect git provider from URL: %s. Supported providers are: github.com, gitlab.com, bitbucket.org", url)
	}
}
