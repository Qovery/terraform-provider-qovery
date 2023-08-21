package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure jobService defined types fully satisfy the job.Service interface.
var _ job.Service = jobService{}

// jobService implements the interface job.Service.
type jobService struct {
	jobRepository        job.Repository
	jobDeploymentService deployment.Service
	variableService      variable.Service
	secretService        secret.Service
}

// NewJobService return a new instance of a job.Service that uses the given job.Repository.
func NewJobService(jobRepository job.Repository, jobDeploymentService deployment.Service, variableService variable.Service, secretService secret.Service) (job.Service, error) {
	if jobRepository == nil {
		return nil, ErrInvalidRepository
	}

	if jobDeploymentService == nil {
		return nil, ErrInvalidService
	}

	if variableService == nil {
		return nil, ErrInvalidService
	}

	if secretService == nil {
		return nil, ErrInvalidService
	}

	return &jobService{
		jobRepository:        jobRepository,
		variableService:      variableService,
		secretService:        secretService,
		jobDeploymentService: jobDeploymentService,
	}, nil
}

// Create handles the domain logic to create a job.
func (s jobService) Create(ctx context.Context, environmentID string, request job.UpsertServiceRequest) (*job.Job, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	newJob, err := s.jobRepository.Create(ctx, environmentID, request.JobUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	overridesAuthorizedScopes := make(map[variable.Scope]struct{})
	overridesAuthorizedScopes[variable.ScopeProject] = struct{}{}
	overridesAuthorizedScopes[variable.ScopeEnvironment] = struct{}{}
	_, err = s.variableService.Update(ctx, newJob.ID.String(), request.EnvironmentVariables, request.EnvironmentVariableAliases, request.EnvironmentVariableOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	_, err = s.secretService.Update(ctx, newJob.ID.String(), request.Secrets, request.SecretAliases, request.SecretOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	newJob, err = s.refreshJob(ctx, *newJob)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToCreateJob.Error())
	}

	return newJob, nil
}

// Get handles the domain logic to retrieve a job.
func (s jobService) Get(ctx context.Context, jobID string) (*job.Job, error) {
	if err := s.checkID(jobID); err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToGetJob.Error())
	}

	fetchedJob, err := s.jobRepository.Get(ctx, jobID)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToGetJob.Error())
	}

	fetchedJob, err = s.refreshJob(ctx, *fetchedJob)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToGetJob.Error())
	}

	return fetchedJob, nil
}

// Update handles the domain logic to update a job.
func (s jobService) Update(ctx context.Context, jobID string, request job.UpsertServiceRequest) (*job.Job, error) {
	if err := s.checkID(jobID); err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	updateJob, err := s.jobRepository.Update(ctx, jobID, request.JobUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	overridesAuthorizedScopes := make(map[variable.Scope]struct{})
	overridesAuthorizedScopes[variable.ScopeProject] = struct{}{}
	overridesAuthorizedScopes[variable.ScopeEnvironment] = struct{}{}
	_, err = s.variableService.Update(ctx, updateJob.ID.String(), request.EnvironmentVariables, request.EnvironmentVariableAliases, request.EnvironmentVariableOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	_, err = s.secretService.Update(ctx, updateJob.ID.String(), request.Secrets, request.SecretAliases, request.SecretOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	updateJob, err = s.refreshJob(ctx, *updateJob)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrFailedToUpdateJob.Error())
	}

	return updateJob, nil
}

// Delete handles the domain logic to delete a job.
func (s jobService) Delete(ctx context.Context, jobID string) error {
	if err := s.checkID(jobID); err != nil {
		return errors.Wrap(err, job.ErrFailedToDeleteJob.Error())
	}

	if err := s.jobRepository.Delete(ctx, jobID); err != nil {
		return errors.Wrap(err, job.ErrFailedToDeleteJob.Error())
	}

	if err := wait(ctx, waitNotFoundFunc(s.jobDeploymentService, jobID), nil); err != nil {
		return errors.Wrap(err, job.ErrFailedToDeleteJob.Error())
	}

	return nil
}

func (s jobService) refreshJob(ctx context.Context, job job.Job) (*job.Job, error) {
	envVars, err := s.variableService.List(ctx, job.ID.String())
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretService.List(ctx, job.ID.String())
	if err != nil {
		return nil, err
	}

	status, err := s.jobDeploymentService.GetStatus(ctx, job.ID.String())
	if err != nil {
		return nil, err
	}

	if err := job.SetEnvironmentVariables(envVars); err != nil {
		return nil, err
	}

	if err := job.SetSecrets(secrets); err != nil {
		return nil, err
	}

	if err := job.SetState(status.State); err != nil {
		return nil, err
	}

	return &job, err
}

// checkEnvironmentID validates that the given environmentID is valid.
func (s jobService) checkEnvironmentID(environmentID string) error {
	if environmentID == "" {
		return job.ErrInvalidJobEnvironmentIDParam
	}

	if _, err := uuid.Parse(environmentID); err != nil {
		return errors.Wrap(err, job.ErrInvalidJobEnvironmentIDParam.Error())
	}

	return nil
}

// checkID validates that the given jobID is valid.
func (s jobService) checkID(jobID string) error {
	if jobID == "" {
		return job.ErrInvalidJobIDParam
	}

	if _, err := uuid.Parse(jobID); err != nil {
		return errors.Wrap(err, job.ErrInvalidJobIDParam.Error())
	}

	return nil
}
