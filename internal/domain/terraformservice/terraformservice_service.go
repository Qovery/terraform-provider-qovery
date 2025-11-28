package terraformservice

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateTerraformService = errors.New("failed to create terraform service")
	ErrFailedToGetTerraformService    = errors.New("failed to get terraform service")
	ErrFailedToUpdateTerraformService = errors.New("failed to update terraform service")
	ErrFailedToDeleteTerraformService = errors.New("failed to delete terraform service")
	ErrFailedToListTerraformServices  = errors.New("failed to list terraform services")
)

// Service represents the interface to implement to handle the domain logic of a TerraformService.
type Service interface {
	Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*TerraformService, error)
	Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*TerraformService, error)
	Update(ctx context.Context, terraformServiceID string, request UpsertServiceRequest) (*TerraformService, error)
	Delete(ctx context.Context, terraformServiceID string) error
	List(ctx context.Context, environmentID string) ([]TerraformService, error)
}

// UpsertServiceRequest represents the parameters needed to create & update a TerraformService.
type UpsertServiceRequest struct {
	TerraformServiceUpsertRequest UpsertRepositoryRequest
	// Future: Add variable/secret diff requests if needed
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) Validate() error {
	if err := r.TerraformServiceUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
