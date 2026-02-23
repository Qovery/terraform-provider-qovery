package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
)

type ApplicationResponse struct {
	ApplicationResponse                     *qovery.Application
	ApplicationDeploymentStageID            string
	ApplicationIsSkipped                    bool
	ApplicationEnvironmentVariables         []*qovery.EnvironmentVariable
	ApplicationEnvironmentVariableAliases   []*qovery.EnvironmentVariable
	ApplicationEnvironmentVariableOverrides []*qovery.EnvironmentVariable
	ApplicationSecrets                      []*qovery.Secret
	ApplicationSecretAliases                []*qovery.Secret
	ApplicationSecretOverrides              []*qovery.Secret
	ApplicationCustomDomains                []*qovery.CustomDomain
	ApplicationDeploymentRestrictions       []deploymentrestriction.ServiceDeploymentRestriction
	ApplicationExternalHost                 *string
	ApplicationInternalHost                 string
	AdvancedSettingsJson                    string
}

type ApplicationCreateParams struct {
	ApplicationRequest               qovery.ApplicationRequest
	ApplicationDeploymentStageID     string
	ApplicationIsSkipped             bool
	EnvironmentVariablesDiff         EnvironmentVariablesDiff
	EnvironmentVariableAliasesDiff   EnvironmentVariablesDiff
	EnvironmentVariableOverridesDiff EnvironmentVariablesDiff
	CustomDomainsDiff                CustomDomainsDiff
	SecretsDiff                      SecretsDiff
	SecretAliasesDiff                SecretsDiff
	SecretOverridesDiff              SecretsDiff
	AdvancedSettingsJson             string
	DeploymentRestrictionsDiff       deploymentrestriction.ServiceDeploymentRestrictionsDiff
}

type ApplicationUpdateParams struct {
	ApplicationEditRequest           qovery.ApplicationEditRequest
	ApplicationDeploymentStageID     string
	ApplicationIsSkipped             bool
	EnvironmentVariablesDiff         EnvironmentVariablesDiff
	EnvironmentVariableAliasesDiff   EnvironmentVariablesDiff
	EnvironmentVariableOverridesDiff EnvironmentVariablesDiff
	CustomDomainsDiff                CustomDomainsDiff
	SecretsDiff                      SecretsDiff
	SecretAliasesDiff                SecretsDiff
	SecretOverridesDiff              SecretsDiff
	AdvancedSettingsJson             string
	DeploymentRestrictionsDiff       deploymentrestriction.ServiceDeploymentRestrictionsDiff
	DockerTargetBuildStage           *string
}

func (c *Client) CreateApplication(ctx context.Context, environmentID string, params *ApplicationCreateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationsAPI.
		CreateApplication(ctx, environmentID).
		ApplicationRequest(params.ApplicationRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, params.ApplicationRequest.Name, res, err)
	}

	// Attach application to deployment stage
	if len(params.ApplicationDeploymentStageID) > 0 {
		attachRequest := qovery.NewAttachServiceToDeploymentStageRequest()
		attachRequest.SetIsSkipped(params.ApplicationIsSkipped)
		_, resp, err := c.api.DeploymentStageMainCallsAPI.
			AttachServiceToDeploymentStage(ctx, params.ApplicationDeploymentStageID, application.Id).
			AttachServiceToDeploymentStageRequest(*attachRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceUpdateDeploymentStage, params.ApplicationDeploymentStageID, resp, err)
		}
	}

	// Get application deployment stage
	applicationDeploymentStage, resp, err := c.api.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, application.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, application.Id, resp, err)
	}

	return c.updateApplication(
		ctx,
		application,
		params.EnvironmentVariablesDiff,
		params.EnvironmentVariableAliasesDiff,
		params.EnvironmentVariableOverridesDiff,
		params.SecretsDiff,
		params.SecretAliasesDiff,
		params.SecretOverridesDiff,
		params.CustomDomainsDiff,
		params.DeploymentRestrictionsDiff,
		applicationDeploymentStage.Id,
		getIsSkippedFromDeploymentStage(applicationDeploymentStage, application.Id),
		params.AdvancedSettingsJson,
	)
}

func (c *Client) GetApplication(ctx context.Context, applicationID string, advancedSettingsFromState string, isTriggeredFromImport bool) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsAPI.
		GetApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
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

	deploymentStage, resp, err := c.api.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, application.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.api.GetConfig()).ReadServiceAdvancedSettings(domain.APPLICATION, applicationID, advancedSettingsFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, nil, err)
	}

	deploymentRestrictionService, err := deploymentrestriction.NewDeploymentRestrictionService(*c.api)
	if err != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, application.Id, nil, err)
	}
	deploymentRestrictions, apiErr := deploymentRestrictionService.GetServiceDeploymentRestrictions(ctx, application.Id, domain.APPLICATION)
	if apiErr != nil {
		return nil, apiErr
	}

	variables := computeAliasOverrideValueVariablesAndSecrets(environmentVariables, secrets)

	return &ApplicationResponse{
		ApplicationResponse:                     application,
		ApplicationDeploymentStageID:            deploymentStage.Id,
		ApplicationIsSkipped:                    getIsSkippedFromDeploymentStage(deploymentStage, application.Id),
		ApplicationEnvironmentVariables:         variables.variableValues,
		ApplicationEnvironmentVariableAliases:   variables.variableAliases,
		ApplicationEnvironmentVariableOverrides: variables.variableOverrides,
		ApplicationSecrets:                      variables.secretValues,
		ApplicationSecretAliases:                variables.secretAliases,
		ApplicationSecretOverrides:              variables.secretOverrides,
		ApplicationCustomDomains:                customDomains,
		ApplicationExternalHost:                 hosts.external,
		ApplicationInternalHost:                 hosts.internal,
		AdvancedSettingsJson:                    *advancedSettingsAsJson,
		ApplicationDeploymentRestrictions:       deploymentRestrictions,
	}, nil
}

func (c *Client) UpdateApplication(ctx context.Context, applicationID string, params *ApplicationUpdateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsAPI.
		EditApplication(ctx, applicationID).
		ApplicationEditRequest(params.ApplicationEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	// Attach service to deployment stage
	if len(params.ApplicationDeploymentStageID) > 0 {
		attachRequest := qovery.NewAttachServiceToDeploymentStageRequest()
		attachRequest.SetIsSkipped(params.ApplicationIsSkipped)
		_, resp, err := c.api.DeploymentStageMainCallsAPI.
			AttachServiceToDeploymentStage(ctx, params.ApplicationDeploymentStageID, applicationID).
			AttachServiceToDeploymentStageRequest(*attachRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceUpdateDeploymentStage, applicationID, resp, err)
		}
	}

	// Get application deployment stage to read IsSkipped
	applicationDeploymentStage, resp, err := c.api.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, applicationID).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, applicationID, resp, err)
	}

	return c.updateApplication(
		ctx,
		application,
		params.EnvironmentVariablesDiff,
		params.EnvironmentVariableAliasesDiff,
		params.EnvironmentVariableOverridesDiff,
		params.SecretsDiff,
		params.SecretAliasesDiff,
		params.SecretOverridesDiff,
		params.CustomDomainsDiff,
		params.DeploymentRestrictionsDiff,
		applicationDeploymentStage.Id,
		getIsSkippedFromDeploymentStage(applicationDeploymentStage, applicationID),
		params.AdvancedSettingsJson,
	)
}

func (c *Client) DeleteApplication(ctx context.Context, applicationID string) *apierrors.APIError {
	application, res, err := c.api.ApplicationMainCallsAPI.
		GetApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		if res.StatusCode == 404 {
			// if the application is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	envChecker := newEnvironmentFinalStateCheckerWaitFunc(c, application.Environment.Id)
	if apiErr := wait(ctx, envChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err = c.api.ApplicationMainCallsAPI.
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

func (c *Client) updateApplication(
	ctx context.Context,
	application *qovery.Application,
	environmentVariablesDiff EnvironmentVariablesDiff,
	environmentVariableAliasesDiff EnvironmentVariablesDiff,
	environmentVariableOverridesDiff EnvironmentVariablesDiff,
	secretsDiff SecretsDiff,
	secretAliasesDiff SecretsDiff,
	secretOverridesDiff SecretsDiff,
	customDomainsDiff CustomDomainsDiff,
	deploymentRestrictionsDiff deploymentrestriction.ServiceDeploymentRestrictionsDiff,
	deploymentStageId string,
	isSkipped bool,
	advancedSettingsJson string,
) (*ApplicationResponse, *apierrors.APIError) {
	if !environmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateApplicationEnvironmentVariables(ctx, application.Id, environmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !environmentVariableAliasesDiff.IsEmpty() || !environmentVariableOverridesDiff.IsEmpty() {
		// For overrides and aliases, we need to retrieve all VALUE secret ids
		variablesByNameForAliases, variablesByNameForOverrides, apiError := c.fetchVariablesForAliasesAndOverrides(ctx, application)
		if apiError != nil {
			return nil, apiError
		}
		// update aliases
		if !environmentVariableAliasesDiff.IsEmpty() {
			if apiErr := c.updateApplicationEnvironmentVariableAliases(ctx, application.Id, environmentVariableAliasesDiff, variablesByNameForAliases); apiErr != nil {
				return nil, apiErr
			}
		}

		// update overrides
		if !environmentVariableOverridesDiff.IsEmpty() {
			if apiErr := c.updateApplicationEnvironmentVariableOverrides(ctx, application.Id, environmentVariableOverridesDiff, variablesByNameForOverrides); apiErr != nil {
				return nil, apiErr
			}
		}
	}

	if !secretsDiff.IsEmpty() {
		if apiErr := c.updateApplicationSecrets(ctx, application.Id, secretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !secretAliasesDiff.IsEmpty() || !secretOverridesDiff.IsEmpty() {
		// For overrides and aliases, we need to retrieve all VALUE secret ids
		secretsByNameForAliases, secretsByNameForOverrides, apiError := c.fetchSecretsForAliasesAndOverrides(ctx, application)
		if apiError != nil {
			return nil, apiError
		}
		// update aliases
		if !secretAliasesDiff.IsEmpty() {
			if apiErr := c.updateApplicationSecretAliases(ctx, application.Id, secretAliasesDiff, secretsByNameForAliases); apiErr != nil {
				return nil, apiErr
			}
		}

		// update overrides
		if !secretOverridesDiff.IsEmpty() {
			if apiErr := c.updateApplicationSecretOverrides(ctx, application.Id, secretOverridesDiff, secretsByNameForOverrides); apiErr != nil {
				return nil, apiErr
			}
		}
	}

	if !customDomainsDiff.IsEmpty() {
		if apiErr := c.updateApplicationCustomDomains(ctx, application.Id, customDomainsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	deploymentRestrictionService, err := deploymentrestriction.NewDeploymentRestrictionService(*c.api)
	if err != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, application.Id, nil, err)
	}
	if deploymentRestrictionsDiff.IsNotEmpty() {
		if apiErr := deploymentRestrictionService.UpdateServiceDeploymentRestrictions(ctx, application.Id, domain.APPLICATION, deploymentRestrictionsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	err = advanced_settings.NewServiceAdvancedSettingsService(c.api.GetConfig()).UpdateServiceAdvancedSettings(domain.APPLICATION, application.Id, advancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, application.Id, nil, err)
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

	deploymentRestrictions, apiErr := deploymentRestrictionService.GetServiceDeploymentRestrictions(ctx, application.Id, domain.APPLICATION)
	if apiErr != nil {
		return nil, apiErr
	}

	variables := computeAliasOverrideValueVariablesAndSecrets(environmentVariables, secrets)
	return &ApplicationResponse{
		ApplicationResponse:                     application,
		ApplicationEnvironmentVariables:         variables.variableValues,
		ApplicationEnvironmentVariableAliases:   variables.variableAliases,
		ApplicationEnvironmentVariableOverrides: variables.variableOverrides,
		ApplicationSecrets:                      variables.secretValues,
		ApplicationSecretAliases:                variables.secretAliases,
		ApplicationSecretOverrides:              variables.secretOverrides,
		ApplicationCustomDomains:                customDomains,
		ApplicationExternalHost:                 hosts.external,
		ApplicationInternalHost:                 hosts.internal,
		ApplicationDeploymentStageID:            deploymentStageId,
		ApplicationIsSkipped:                    isSkipped,
		AdvancedSettingsJson:                    advancedSettingsJson,
		ApplicationDeploymentRestrictions:       deploymentRestrictions,
	}, nil
}

// getIsSkippedFromDeploymentStage returns is_skipped for a service within a deployment stage response.
func getIsSkippedFromDeploymentStage(deploymentStage *qovery.DeploymentStageResponse, serviceID string) bool {
	if deploymentStage == nil {
		return false
	}
	for _, svc := range deploymentStage.GetServices() {
		if svc.GetServiceId() == serviceID {
			return svc.GetIsSkipped()
		}
	}
	return false
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
		if env.Key == hostExternalKey && env.Value != nil {
			hostExternal = *env.Value
			continue
		}
		if env.Key == hostInternalKey && env.Value != nil {
			hostInternal = *env.Value
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

// fetchVariablesForAliasesAndOverrides
// returns 2 hashmaps used to send requests for variable aliases & overrides
func (c *Client) fetchVariablesForAliasesAndOverrides(ctx context.Context, app *qovery.Application) (map[string]qovery.EnvironmentVariable, map[string]qovery.EnvironmentVariable, *apierrors.APIError) {
	applicationVariables, response, err := c.api.ApplicationEnvironmentVariableAPI.ListApplicationEnvironmentVariable(ctx, app.Id).Execute()
	if err != nil || response.StatusCode >= 400 {
		return nil, nil, apierrors.NewReadError(apierrors.APIResourceApplicationEnvironmentVariable, app.Id, response, err)
	}

	variablesByNameForAliases := make(map[string]qovery.EnvironmentVariable)
	variablesByNameForOverrides := make(map[string]qovery.EnvironmentVariable)
	for _, result := range applicationVariables.Results {
		if result.VariableType == qovery.APIVARIABLETYPEENUM_VALUE || result.VariableType == qovery.APIVARIABLETYPEENUM_BUILT_IN {
			variablesByNameForAliases[result.Key] = result
		}
		if result.VariableType == qovery.APIVARIABLETYPEENUM_VALUE && (result.Scope == qovery.APIVARIABLESCOPEENUM_ENVIRONMENT || result.Scope == qovery.APIVARIABLESCOPEENUM_PROJECT) {
			variablesByNameForOverrides[result.Key] = result
		}
	}

	return variablesByNameForAliases, variablesByNameForOverrides, nil
}

// fetchSecretsForAliasesAndOverrides
// returns 2 hashmaps used to send requests for secret aliases & overrides
func (c *Client) fetchSecretsForAliasesAndOverrides(ctx context.Context, app *qovery.Application) (map[string]qovery.Secret, map[string]qovery.Secret, *apierrors.APIError) {
	applicationVariables, response, err := c.api.ApplicationSecretAPI.ListApplicationSecrets(ctx, app.Id).Execute()
	if err != nil || response.StatusCode >= 400 {
		return nil, nil, apierrors.NewReadError(apierrors.APIResourceApplicationSecret, app.Id, response, err)
	}

	secretsByNameForAliases := make(map[string]qovery.Secret)
	secretsByNameForOverrides := make(map[string]qovery.Secret)
	for _, result := range applicationVariables.Results {
		if *result.VariableType == qovery.APIVARIABLETYPEENUM_VALUE || *result.VariableType == qovery.APIVARIABLETYPEENUM_BUILT_IN {
			secretsByNameForAliases[result.Key] = result
		}
		if *result.VariableType == qovery.APIVARIABLETYPEENUM_VALUE && (result.Scope == qovery.APIVARIABLESCOPEENUM_ENVIRONMENT || result.Scope == qovery.APIVARIABLESCOPEENUM_PROJECT) {
			secretsByNameForOverrides[result.Key] = result
		}
	}

	return secretsByNameForAliases, secretsByNameForOverrides, nil
}

type ValueAliasOverrideApplicationVariable struct {
	variableValues    []*qovery.EnvironmentVariable
	variableAliases   []*qovery.EnvironmentVariable
	variableOverrides []*qovery.EnvironmentVariable
	secretValues      []*qovery.Secret
	secretAliases     []*qovery.Secret
	secretOverrides   []*qovery.Secret
}

func computeAliasOverrideValueVariablesAndSecrets(
	environmentVariables []*qovery.EnvironmentVariable,
	secrets []*qovery.Secret,
) ValueAliasOverrideApplicationVariable {
	// We need to create 3 different lists from all variables to satisfy terraform attributes: VALUE / ALIAS / OVERRIDE
	var variableValues []*qovery.EnvironmentVariable
	var variableAliases []*qovery.EnvironmentVariable
	var variableOverrides []*qovery.EnvironmentVariable

	for _, variable := range environmentVariables {
		if variable.VariableType == qovery.APIVARIABLETYPEENUM_VALUE || variable.VariableType == qovery.APIVARIABLETYPEENUM_BUILT_IN {
			variableValues = append(variableValues, variable)
		}
		if variable.VariableType == qovery.APIVARIABLETYPEENUM_ALIAS {
			variableAliases = append(variableAliases, variable)
		}
		if variable.VariableType == qovery.APIVARIABLETYPEENUM_OVERRIDE {
			variableOverrides = append(variableOverrides, variable)
		}
	}

	// We need to create 3 different lists from all secrets to satisfy terraform attributes: VALUE / ALIAS / OVERRIDE
	var secretValues []*qovery.Secret
	var secretAliases []*qovery.Secret
	var secretOverrides []*qovery.Secret

	for _, secret := range secrets {
		if *secret.VariableType == qovery.APIVARIABLETYPEENUM_VALUE || *secret.VariableType == qovery.APIVARIABLETYPEENUM_BUILT_IN {
			secretValues = append(secretValues, secret)
		}
		if *secret.VariableType == qovery.APIVARIABLETYPEENUM_ALIAS {
			secretAliases = append(secretAliases, secret)
		}
		if *secret.VariableType == qovery.APIVARIABLETYPEENUM_OVERRIDE {
			secretOverrides = append(secretOverrides, secret)
		}
	}

	return ValueAliasOverrideApplicationVariable{
		variableValues:    variableValues,
		variableAliases:   variableAliases,
		variableOverrides: variableOverrides,
		secretValues:      secretValues,
		secretAliases:     secretAliases,
		secretOverrides:   secretOverrides,
	}
}
