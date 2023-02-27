package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type ApplicationResponse struct {
	ApplicationResponse             *qovery.Application
	ApplicationDeploymentStageId    string
	ApplicationStatus               *qovery.Status
	ApplicationEnvironmentVariables []*qovery.EnvironmentVariable
	ApplicationSecrets              []*qovery.Secret
	ApplicationCustomDomains        []*qovery.CustomDomain
	ApplicationExternalHost         *string
	ApplicationInternalHost         string
}

type ApplicationCreateParams struct {
	ApplicationRequest           qovery.ApplicationRequest
	ApplicationDeploymentStageId string
	EnvironmentVariablesDiff     EnvironmentVariablesDiff
	CustomDomainsDiff            CustomDomainsDiff
	SecretsDiff                  SecretsDiff
	DesiredState                 qovery.StateEnum
}

type ApplicationUpdateParams struct {
	ApplicationEditRequest       qovery.ApplicationEditRequest
	ApplicationDeploymentStageId string
	EnvironmentVariablesDiff     EnvironmentVariablesDiff
	CustomDomainsDiff            CustomDomainsDiff
	SecretsDiff                  SecretsDiff
	DesiredState                 qovery.StateEnum
}

func (c *Client) CreateApplication(ctx context.Context, environmentID string, params *ApplicationCreateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationsApi.
		CreateApplication(ctx, environmentID).
		ApplicationRequest(params.ApplicationRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, params.ApplicationRequest.Name, res, err)
	}

	// Attach service to deployment stage
	if len(params.ApplicationDeploymentStageId) > 0 {
		_, resp, err := c.api.DeploymentStageMainCallsApi.
			AttachServiceToDeploymentStage(ctx, params.ApplicationDeploymentStageId, application.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceUpdateDeploymentStage, application.Id, resp, err)
		}
	}

	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.SecretsDiff, params.CustomDomainsDiff, params.DesiredState, params.ApplicationDeploymentStageId)
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

	hosts, apiErr := c.getApplicationHosts(ctx, application, environmentVariables)
	if apiErr != nil {
		return nil, apiErr
	}

	// TODO (mzo) use deployment_stage_id available in ApplicationResponse when it is merged
	deploymentStages, response, err := c.api.DeploymentStageMainCallsApi.ListEnvironmentDeploymentStage(ctx, application.Environment.Id).Execute()
	if err != nil || response.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}
	var deploymentStageId = ""
	for _, deploymentStage := range deploymentStages.GetResults() {
		for _, deploymentStageService := range deploymentStage.Services {
			if deploymentStageService.ServiceId == &applicationID {
				deploymentStageId = deploymentStage.Id
				break
			}
		}
	}
	// END TODO (mzo)

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationDeploymentStageId:    deploymentStageId,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
		ApplicationSecrets:              secrets,
		ApplicationCustomDomains:        customDomains,
		ApplicationExternalHost:         hosts.external,
		ApplicationInternalHost:         hosts.internal,
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

	// Attach service to deployment stage
	if len(params.ApplicationDeploymentStageId) > 0 {
		_, resp, err := c.api.DeploymentStageMainCallsApi.
			AttachServiceToDeploymentStage(ctx, params.ApplicationDeploymentStageId, applicationID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceUpdateDeploymentStage, applicationID, resp, err)
		}
	}

	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.SecretsDiff, params.CustomDomainsDiff, params.DesiredState, params.ApplicationDeploymentStageId)
}

func (c *Client) DeleteApplication(ctx context.Context, applicationID string) *apierrors.APIError {
	application, apiErr := c.GetApplication(ctx, applicationID)
	if apiErr != nil {
		return apiErr
	}

	envChecker := newEnvironmentFinalStateCheckerWaitFunc(c, application.ApplicationResponse.Environment.Id)
	if apiErr := wait(ctx, envChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.api.ApplicationMainCallsApi.
		DeleteApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	checker := newApplicationStatusCheckerWaitFunc(c, applicationID, qovery.STATEENUM_DELETED)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) updateApplication(ctx context.Context, application *qovery.Application, environmentVariablesDiff EnvironmentVariablesDiff, secretsDiff SecretsDiff, customDomainsDiff CustomDomainsDiff, desiredState qovery.StateEnum, deploymentStageId string) (*ApplicationResponse, *apierrors.APIError) {
	forceRedeploy := !environmentVariablesDiff.IsEmpty() || !secretsDiff.IsEmpty() || !customDomainsDiff.IsEmpty()
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

	status, apiErr := c.updateApplicationStatus(ctx, application, desiredState, forceRedeploy)
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

	hosts, apiErr := c.getApplicationHosts(ctx, application, environmentVariables)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
		ApplicationSecrets:              secrets,
		ApplicationCustomDomains:        customDomains,
		ApplicationExternalHost:         hosts.external,
		ApplicationInternalHost:         hosts.internal,
		ApplicationDeploymentStageId:    deploymentStageId,
	}, nil
}

type applicationHosts struct {
	internal string
	external *string
}

func (c *Client) getApplicationHosts(ctx context.Context, application *qovery.Application, environmentVariables []*qovery.EnvironmentVariable) (*applicationHosts, *apierrors.APIError) {
	// Get all environment variables associated to this application,
	// and pick only the elements that I need to construct my struct below
	// Context: since I need to get the internal host of my application and this information is only available via the environment env vars,
	// then we list all env vars from the environment where the application is to take it.
	// FIXME - it's a really bad idea of doing that but I have no choice... If we change the way we structure environment variable backend side, then we will be f***ed up :/
	hostExternalKey := fmt.Sprintf("QOVERY_APPLICATION_Z%s_HOST_EXTERNAL", strings.ToUpper(strings.Split(application.Id, "-")[0]))
	hostInternalKey := fmt.Sprintf("QOVERY_APPLICATION_Z%s_HOST_INTERNAL", strings.ToUpper(strings.Split(application.Id, "-")[0]))
	// Expected host external key syntax is `QOVERY_APPLICATION_Z{APP-ID}_HOST_EXTERNAL`
	// Expected host internal key syntax is `QOVERY_APPLICATION_Z{APP-ID}_HOST_INTERNAL`

	hostExternal := ""
	hostInternal := ""
	for _, env := range environmentVariables {
		if env.Key == hostExternalKey {
			hostExternal = env.Value
			continue
		}
		if env.Key == hostInternalKey {
			hostInternal = env.Value
			continue
		}
		if hostInternal != "" && hostExternal != "" {
			break
		}
	}

	hosts := &applicationHosts{
		internal: hostInternal,
	}
	if hostExternal != "" {
		hosts.external = &hostExternal
	}

	return hosts, nil
}
