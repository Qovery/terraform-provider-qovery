package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure projectQoveryAPI defined types fully satisfy the project.Repository interface.
var _ variable.Repository = projectEnvironmentVariablesQoveryAPI{}

// projectQoveryAPI implements the interface project.Repository.
type projectQoveryAPI struct {
	client *qovery.APIClient
}

// newProjectQoveryAPI return a new instance of a project.Repository that uses Qovery's API.
func newProjectQoveryAPI(client *qovery.APIClient) (project.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &projectQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a project for an organization using the given organizationID and request.
func (c projectQoveryAPI) Create(ctx context.Context, organizationID string, request project.UpsertRepositoryRequest) (*project.Project, error) {
	proj, resp, err := c.client.ProjectsApi.
		CreateProject(ctx, organizationID).
		ProjectRequest(newQoveryProjectRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceProject, request.Name, resp, err)
	}

	return newDomainProjectFromQovery(proj)
}

// Get calls Qovery's API to retrieve a  project using the given projectID.
func (c projectQoveryAPI) Get(ctx context.Context, projectID string) (*project.Project, error) {
	proj, resp, err := c.client.ProjectMainCallsApi.
		GetProject(ctx, projectID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceProject, projectID, resp, err)
	}

	return newDomainProjectFromQovery(proj)
}

// Update calls Qovery's API to update a project using the given projectID and request.
func (c projectQoveryAPI) Update(ctx context.Context, projectID string, request project.UpsertRepositoryRequest) (*project.Project, error) {
	proj, resp, err := c.client.ProjectMainCallsApi.
		EditProject(ctx, projectID).
		ProjectRequest(newQoveryProjectRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceProject, projectID, resp, err)
	}

	return newDomainProjectFromQovery(proj)
}

// Delete calls Qovery's API to deletes a project using the given projectID.
func (c projectQoveryAPI) Delete(ctx context.Context, projectID string) error {
	resp, err := c.client.ProjectMainCallsApi.
		DeleteProject(ctx, projectID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceProject, projectID, resp, err)
	}

	return nil
}
