package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
)

// environmentQoveryAPI implements the interface environment.Repository.
type environmentQoveryAPI struct {
	client *qovery.APIClient
}

// newEnvironmentQoveryAPI return a new instance of an environment.Repository that uses Qovery's API.
func newEnvironmentQoveryAPI(client *qovery.APIClient) (environment.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &environmentQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment for an organization using the given projectID and request.
func (c environmentQoveryAPI) Create(ctx context.Context, projectID string, request environment.CreateRepositoryRequest) (*environment.Environment, error) {
	req, err := newQoveryCreateEnvironmentRequestFromDomain(request)
	if err != nil {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceEnvironment, request.Name, nil, err)
	}

	env, resp, err := c.client.EnvironmentsAPI.
		CreateEnvironment(ctx, projectID).
		CreateEnvironmentRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceEnvironment, request.Name, resp, err)
	}

	return newDomainEnvironmentFromQovery(env)
}

// Get calls Qovery's API to retrieve an environment using the given environmentID.
func (c environmentQoveryAPI) Get(ctx context.Context, environmentID string) (*environment.Environment, error) {
	env, resp, err := c.client.EnvironmentMainCallsAPI.
		GetEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceEnvironment, environmentID, resp, err)
	}

	return newDomainEnvironmentFromQovery(env)
}

// Update calls Qovery's API to update an environment using the given environmentID and request.
func (c environmentQoveryAPI) Update(ctx context.Context, environmentID string, request environment.UpdateRepositoryRequest) (*environment.Environment, error) {
	req, err := newQoveryEnvironmentEditRequestFromDomain(request)
	if err != nil {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceEnvironment, environmentID, nil, err)
	}

	env, resp, err := c.client.EnvironmentMainCallsAPI.
		EditEnvironment(ctx, environmentID).
		EnvironmentEditRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceEnvironment, environmentID, resp, err)
	}

	return newDomainEnvironmentFromQovery(env)
}

// Delete calls Qovery's API to deletes an environment using the given environmentID.
func (c environmentQoveryAPI) Delete(ctx context.Context, environmentID string) error {
	resp, err := c.client.EnvironmentMainCallsAPI.
		DeleteEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironment, environmentID, resp, err)
	}

	return nil
}

func (c environmentQoveryAPI) Exists(ctx context.Context, environmentID string) bool {
	_, resp, _ := c.client.EnvironmentMainCallsAPI.
		GetEnvironment(ctx, environmentID).
		Execute()
	return !(resp.StatusCode >= 400)
}
