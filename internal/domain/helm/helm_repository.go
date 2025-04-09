package helm

//go:generate mockery --testonly --with-expecter --name=Repository --structname=HelmRepository --filename=helm_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*Helm, error)
	Get(ctx context.Context, helmID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*Helm, error)
	Update(ctx context.Context, helmID string, request UpsertRepositoryRequest) (*Helm, error)
	Delete(ctx context.Context, helmID string) error
}

type UpsertRepositoryRequest struct {
	Name                      string `validate:"required"`
	Description               *string
	IconUri                   *string
	TimeoutSec                *int32
	AutoPreview               qovery.NullableBool
	AutoDeploy                bool
	Arguments                 []string
	AllowClusterWideResources bool
	Source                    Source
	ValuesOverride            ValuesOverride
	Ports                     *[]Port
	EnvironmentVariables      []variable.UpsertRequest
	Secrets                   []secret.UpsertRequest
	DeploymentStageID         string
	AdvancedSettingsJson      string
	CustomDomains             client.CustomDomainsDiff
}

func (r UpsertRepositoryRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidHelmUpsertRequest.Error())
	}

	return nil
}

func (r UpsertRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
