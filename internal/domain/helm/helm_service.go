package helm

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	ErrFailedToCreateHelm = errors.New("failed to create helm")
	ErrFailedToGetHelm    = errors.New("failed to get helm")
	ErrFailedToUpdateHelm = errors.New("failed to update helm")
	ErrFailedToDeleteHelm = errors.New("failed to delete helm")
)

type Service interface {
	Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*Helm, error)
	Get(ctx context.Context, helmID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*Helm, error)
	Update(ctx context.Context, helmID string, request UpsertServiceRequest) (*Helm, error)
	Delete(ctx context.Context, helmID string) error
}

type UpsertServiceRequest struct {
	HelmUpsertRequest            UpsertRepositoryRequest
	EnvironmentVariables         variable.DiffRequest
	EnvironmentVariableAliases   variable.DiffRequest
	EnvironmentVariableOverrides variable.DiffRequest
	Secrets                      secret.DiffRequest
	SecretAliases                secret.DiffRequest
	SecretOverrides              secret.DiffRequest
	DeploymentRestrictionsDiff   deploymentrestriction.ServiceDeploymentRestrictionsDiff
}

func (r UpsertServiceRequest) Validate() error {
	if err := r.HelmUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidHelmUpsertRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidHelmUpsertRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidHelmUpsertRequest.Error())
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidHelmUpsertRequest.Error())
	}

	return nil
}

func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
