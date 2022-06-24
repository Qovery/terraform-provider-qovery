package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type ApplicationResponse struct {
	ApplicationResponse             *qovery.Application
	ApplicationStatus               *qovery.Status
	ApplicationEnvironmentVariables []*qovery.EnvironmentVariable
	ApplicationSecrets              []*qovery.Secret
	ApplicationCustomDomains        []*qovery.CustomDomain
}

type ApplicationCreateParams struct {
	ApplicationRequest       qovery.ApplicationRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	CustomDomainsDiff        CustomDomainsDiff
	SecretsDiff              SecretsDiff
	DesiredState             qovery.StateEnum
}

type ApplicationUpdateParams struct {
	ApplicationEditRequest   qovery.ApplicationEditRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	CustomDomainsDiff        CustomDomainsDiff
	SecretsDiff              SecretsDiff
	DesiredState             qovery.StateEnum
}

func (c *Client) CreateApplication(ctx context.Context, environmentID string, params *ApplicationCreateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationsApi.
		CreateApplication(ctx, environmentID).
		ApplicationRequest(params.ApplicationRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, params.ApplicationRequest.Name, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.SecretsDiff, params.CustomDomainsDiff, params.DesiredState)
}

func (c *Client) GetApplication(ctx context.Context, applicationID string) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsApi.
		GetApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	status, apiErr := c.getApplicationStatus(ctx, applicationID)
	if apiErr != nil {
		return nil, apiErr
	}

	environmentVariables, apiErr := c.getApplicationEnvironmentVariables(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getApplicationSecrets(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	customDomains, apiErr := c.getApplicationCustomDomains(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
		ApplicationSecrets:              secrets,
		ApplicationCustomDomains:        customDomains,
	}, nil
}

func (c *Client) UpdateApplication(ctx context.Context, applicationID string, params *ApplicationUpdateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsApi.
		EditApplication(ctx, applicationID).
		ApplicationEditRequest(params.ApplicationEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, applicationID, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.SecretsDiff, params.CustomDomainsDiff, params.DesiredState)
}

func (c *Client) DeleteApplication(ctx context.Context, applicationID string) *apierrors.APIError {
	finalStateChecker := newApplicationFinalStateCheckerWaitFunc(c, applicationID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.api.ApplicationMainCallsApi.
		DeleteApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	checker := newApplicationStatusCheckerWaitFunc(c, applicationID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) updateApplication(ctx context.Context, application *qovery.Application, environmentVariablesDiff EnvironmentVariablesDiff, secretsDiff SecretsDiff, customDomainsDiff CustomDomainsDiff, desiredState qovery.StateEnum) (*ApplicationResponse, *apierrors.APIError) {
	forceRestart := !environmentVariablesDiff.IsEmpty()
	if !environmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateApplicationEnvironmentVariables(ctx, application.Id, environmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !secretsDiff.IsEmpty() {
		if apiErr := c.updateApplicationSecrets(ctx, application.Id, secretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !customDomainsDiff.IsEmpty() {
		if apiErr := c.updateApplicationCustomDomains(ctx, application.Id, customDomainsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	status, apiErr := c.updateApplicationStatus(ctx, application, desiredState, forceRestart)
	if apiErr != nil {
		return nil, apiErr
	}

	environmentVariables, apiErr := c.getApplicationEnvironmentVariables(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getApplicationSecrets(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	customDomains, apiErr := c.getApplicationCustomDomains(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
		ApplicationSecrets:              secrets,
		ApplicationCustomDomains:        customDomains,
	}, nil
}
