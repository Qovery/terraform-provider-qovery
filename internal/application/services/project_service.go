package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure projectService defined types fully satisfy the project.Service interface.
var _ project.Service = projectService{}

// projectService implements the interface project.Service.
type projectService struct {
	projectRepository project.Repository
	variableService   variable.Service
	secretService     secret.Service
}

// NewProjectService return a new instance of a project.Service that uses the given project.Repository.
func NewProjectService(projectRepository project.Repository, variableService variable.Service, secretService secret.Service) (project.Service, error) {
	if projectRepository == nil {
		return nil, ErrInvalidRepository
	}

	if variableService == nil {
		return nil, ErrInvalidService
	}

	if secretService == nil {
		return nil, ErrInvalidService
	}

	return &projectService{
		projectRepository: projectRepository,
		variableService:   variableService,
		secretService:     secretService,
	}, nil
}

// Create handles the domain logic to create an aws cluster project.
func (s projectService) Create(ctx context.Context, organizationID string, request project.UpsertServiceRequest) (*project.Project, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	proj, err := s.projectRepository.Create(ctx, organizationID, request.ProjectUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	_, err = s.variableService.Update(ctx, proj.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	_, err = s.secretService.Update(ctx, proj.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	proj, err = s.refreshProject(ctx, *proj)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToCreateProject.Error())
	}

	return proj, nil
}

// Get handles the domain logic to retrieve an aws cluster project.
func (s projectService) Get(ctx context.Context, projectID string) (*project.Project, error) {
	if err := s.checkProjectID(projectID); err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToGetProject.Error())
	}

	proj, err := s.projectRepository.Get(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToGetProject.Error())
	}

	proj, err = s.refreshProject(ctx, *proj)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToGetProject.Error())
	}

	return proj, nil
}

// Update handles the domain logic to update an aws cluster project.
func (s projectService) Update(ctx context.Context, projectID string, request project.UpsertServiceRequest) (*project.Project, error) {
	if err := s.checkProjectID(projectID); err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	proj, err := s.projectRepository.Update(ctx, projectID, request.ProjectUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	_, err = s.variableService.Update(ctx, proj.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	_, err = s.secretService.Update(ctx, proj.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	proj, err = s.refreshProject(ctx, *proj)
	if err != nil {
		return nil, errors.Wrap(err, project.ErrFailedToUpdateProject.Error())
	}

	return proj, nil
}

// Delete handles the domain logic to delete an aws cluster project.
func (s projectService) Delete(ctx context.Context, projectID string) error {
	if err := s.checkProjectID(projectID); err != nil {
		return errors.Wrap(err, project.ErrFailedToDeleteProject.Error())
	}

	if err := s.projectRepository.Delete(ctx, projectID); err != nil {
		return errors.Wrap(err, project.ErrFailedToDeleteProject.Error())
	}

	return nil
}

func (s projectService) refreshProject(ctx context.Context, proj project.Project) (*project.Project, error) {
	envVars, err := s.variableService.List(ctx, proj.ID.String())
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretService.List(ctx, proj.ID.String())
	if err != nil {
		return nil, err
	}

	if err := proj.SetEnvironmentVariables(envVars); err != nil {
		return nil, err
	}

	if err := proj.SetSecrets(secrets); err != nil {
		return nil, err
	}

	return &proj, err
}

// checkOrganizationID validates that the given organizationID is valid.
func (s projectService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return project.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, project.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkProjectID validates that the given projectID is valid.
func (s projectService) checkProjectID(projectID string) error {
	if projectID == "" {
		return project.ErrInvalidProjectIDParam
	}

	if _, err := uuid.Parse(projectID); err != nil {
		return errors.Wrap(err, project.ErrInvalidProjectIDParam.Error())
	}

	return nil
}
