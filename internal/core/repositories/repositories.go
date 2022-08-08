package repositories

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/inmem"
	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/qoveryapi"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

var (
	ErrFailedToInitializeQoveryAPI      = errors.New("failed to initialize qovery api")
	ErrMissingRepositoriesConfiguration = errors.New("missing repositories configuration")
)

type Configuration func(repos *Repositories) error

type Repositories struct {
	CredentialsAws      credentials.AwsRepository
	CredentialsScaleway credentials.ScalewayRepository
	Organization        organization.Repository
}

func New(configs ...Configuration) (*Repositories, error) {
	if len(configs) == 0 {
		return nil, ErrMissingRepositoriesConfiguration
	}

	repos := &Repositories{}

	// Apply all the configs to the qoveryAPI instance.
	for _, config := range configs {
		if err := config(repos); err != nil {
			return nil, err
		}
	}

	return repos, nil
}

func WithQoveryAPI(apiToken string, providerVersion string) Configuration {
	return func(repos *Repositories) error {
		qoveryAPI, err := qoveryapi.New(
			qoveryapi.WithQoveryAPIToken(apiToken),
			qoveryapi.WithUserAgent(fmt.Sprintf("terraform-provider-qovery/%s", providerVersion)),
		)
		if err != nil {
			return errors.Wrap(err, ErrFailedToInitializeQoveryAPI.Error())
		}

		repos.CredentialsAws = qoveryAPI.CredentialsAws
		repos.CredentialsScaleway = qoveryAPI.CredentialsScaleway
		repos.Organization = qoveryAPI.Organization

		return nil
	}
}

func WithInmem() Configuration {
	return func(repos *Repositories) error {
		repos.CredentialsAws = inmem.NewCredentialsAwsInmem()
		repos.CredentialsScaleway = inmem.NewCredentialsScalewayInmem()

		return nil
	}
}
