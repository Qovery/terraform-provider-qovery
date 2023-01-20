package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure environmentService defined types fully satisfy the environment.Service interface.
var _ environment.Service = environmentService{}

// environmentService implements the interface environment.Service.
type environmentService struct {
	environmentRepository        environment.Repository
	environmentDeploymentService deployment.Service
	variableService              variable.Service
	secretService                secret.Service
}

// NewEnvironmentService return a new instance of an environment.Service that uses the given environment.Repository.
func NewEnvironmentService(environmentRepository environment.Repository, environmentDeploymentService deployment.Service, variableService variable.Service, secretService secret.Service) (environment.Service, error) {
	if environmentRepository == nil {
		return nil, ErrInvalidRepository
	}

	if environmentDeploymentService == nil {
		return nil, ErrInvalidService
	}

	if variableService == nil {
		return nil, ErrInvalidService
	}

	if secretService == nil {
		return nil, ErrInvalidService
	}

	return &environmentService{
		environmentRepository:        environmentRepository,
		environmentDeploymentService: environmentDeploymentService,
		variableService:              variableService,
		secretService:                secretService,
	}, nil
}

// Create handles the domain logic to create an aws cluster environment.
func (s environmentService) Create(ctx context.Context, projectID string, request environment.CreateServiceRequest) (*environment.Environment, error) {
	if err := s.checkProjectID(projectID); err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	env, err := s.environmentRepository.Create(ctx, projectID, request.EnvironmentCreateRequest)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	_, err = s.variableService.Update(ctx, env.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	_, err = s.secretService.Update(ctx, env.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	env, err = s.refreshEnvironment(ctx, *env)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToCreateEnvironment.Error())
	}

	return env, nil
}

// Get handles the domain logic to retrieve an aws cluster environment.
func (s environmentService) Get(ctx context.Context, environmentID string) (*environment.Environment, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToGetEnvironment.Error())
	}

	env, err := s.environmentRepository.Get(ctx, environmentID)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToGetEnvironment.Error())
	}

	env, err = s.refreshEnvironment(ctx, *env)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToGetEnvironment.Error())
	}

	return env, nil
}

// Update handles the domain logic to update an aws cluster environment.
func (s environmentService) Update(ctx context.Context, environmentID string, request environment.UpdateServiceRequest) (*environment.Environment, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	env, err := s.environmentRepository.Update(ctx, environmentID, request.EnvironmentUpdateRequest)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	_, err = s.variableService.Update(ctx, env.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	_, err = s.secretService.Update(ctx, env.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	if !request.EnvironmentVariables.IsEmpty() || !request.Secrets.IsEmpty() {
		_, err := s.environmentDeploymentService.Redeploy(ctx, environmentID)
		if err != nil {
			return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
		}
	}

	env, err = s.refreshEnvironment(ctx, *env)
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrFailedToUpdateEnvironment.Error())
	}

	return env, nil
}

// Delete handles the domain logic to delete an aws cluster environment.
func (s environmentService) Delete(ctx context.Context, environmentID string) error {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return errors.Wrap(err, environment.ErrFailedToDeleteEnvironment.Error())
	}

	if err := wait(ctx, waitFinalStateFunc(s.environmentDeploymentService, environmentID), nil); err != nil {
		return errors.Wrap(err, environment.ErrFailedToDeleteEnvironment.Error())
	}

	if err := s.environmentRepository.Delete(ctx, environmentID); err != nil {
		return errors.Wrap(err, environment.ErrFailedToDeleteEnvironment.Error())
	}

	if err := wait(ctx, waitNotFoundFunc(s.environmentDeploymentService, environmentID), nil); err != nil {
		return errors.Wrap(err, environment.ErrFailedToDeleteEnvironment.Error())
	}

	return nil
}

func (s environmentService) refreshEnvironment(ctx context.Context, env environment.Environment) (*environment.Environment, error) {
	envVars, err := s.variableService.List(ctx, env.ID.String())
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretService.List(ctx, env.ID.String())
	if err != nil {
		return nil, err
	}

	if err := env.SetEnvironmentVariables(envVars); err != nil {
		return nil, err
	}

	if err := env.SetSecrets(secrets); err != nil {
		return nil, err
	}

	return &env, err
}

// checkProjectID validates that the given projectID is valid.
func (s environmentService) checkProjectID(projectID string) error {
	if projectID == "" {
		return environment.ErrInvalidProjectIDParam
	}

	if _, err := uuid.Parse(projectID); err != nil {
		return errors.Wrap(err, environment.ErrInvalidProjectIDParam.Error())
	}

	return nil
}

// checkEnvironmentID validates that the given environmentID is valid.
func (s environmentService) checkEnvironmentID(environmentID string) error {
	if environmentID == "" {
		return environment.ErrInvalidEnvironmentIDParam
	}

	if _, err := uuid.Parse(environmentID); err != nil {
		return errors.Wrap(err, environment.ErrInvalidEnvironmentIDParam.Error())
	}

	return nil
}
