package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/qoveryapi"
)

// gitProviderFromDomain maps the git provider a domain entity read from the API to the value to
// store, following qoveryapi.GitProviderFromAPI.
func gitProviderFromDomain(url string, provider *string, prior *string) *string {
	if provider == nil {
		return prior
	}
	return qoveryapi.GitProviderFromAPI(url, qovery.GitProviderEnum(*provider), prior)
}
