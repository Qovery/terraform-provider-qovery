package terraformservice

//go:generate mockery --testonly --with-expecter --name=Repository --structname=TerraformServiceRepository --filename=terraformservice_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Repository represents the interface to implement to handle the persistence of a TerraformService.
type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*TerraformService, error)
	Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*TerraformService, error)
	Update(ctx context.Context, terraformServiceID string, request UpsertRepositoryRequest) (*TerraformService, error)
	Delete(ctx context.Context, terraformServiceID string) error
	List(ctx context.Context, environmentID string) ([]TerraformService, error)
}

// UpsertRepositoryRequest represents the parameters needed to create & update a TerraformService.
type UpsertRepositoryRequest struct {
	Name                  string `validate:"required"`
	Description           *string
	AutoDeploy            bool
	DeploymentStageID     string
	IsSkipped             bool
	GitRepository         GitRepository `validate:"required"`
	TfVarFiles            []string
	Variables             []Variable
	Backend               Backend       `validate:"required"`
	Engine                Engine        `validate:"required"`
	EngineVersion         EngineVersion `validate:"required"`
	JobResources          JobResources  `validate:"required"`
	TimeoutSec            *int32
	IconURI               string
	UseClusterCredentials bool
	ActionExtraArguments  map[string][]string
	AdvancedSettingsJson  string
}

// Validate returns an error to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) Validate() error {
	// Validate name
	if r.Name == "" {
		return ErrInvalidTerraformServiceNameParam
	}

	// Validate git repository
	if err := r.GitRepository.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Validate backend
	if err := r.Backend.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Validate engine
	if err := r.Engine.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Validate engine version
	if err := r.EngineVersion.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Validate job resources
	if err := r.JobResources.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Validate variables
	for _, variable := range r.Variables {
		if err := variable.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
		}
	}

	// Validate timeout
	if r.TimeoutSec != nil && *r.TimeoutSec < MinTimeoutSec {
		return errors.Wrap(
			errors.New("timeout_sec must be at least 0"),
			ErrInvalidTerraformServiceUpsertRequest.Error(),
		)
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
