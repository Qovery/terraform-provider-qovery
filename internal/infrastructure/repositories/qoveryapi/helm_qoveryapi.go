package qoveryapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
)

// Ensure helmQoveryAPI defined types fully satisfy the helm.Repository interface.
var _ helm.Repository = helmQoveryAPI{}

// helmQoveryAPI implements the interface helm.Repository.
type helmQoveryAPI struct {
	client *qovery.APIClient
}

// newHelmQoveryAPI return a new instance of a helm.Repository that uses Qovery's API.
func newHelmQoveryAPI(client *qovery.APIClient) (helm.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &helmQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a helm for an organization using the given organizationID and request.
func (c helmQoveryAPI) Create(ctx context.Context, environmentID string, request helm.UpsertRepositoryRequest) (*helm.Helm, error) {
	req, err := newQoveryHelmRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrInvalidHelmUpsertRequest.Error())
	}

	newHelm, resp, err := c.client.HelmsAPI.
		CreateHelm(ctx, environmentID).
		HelmRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, resp, err)
	}

	// The helm service exists in Qovery from here on. partial carries what is already
	// known about it, so that a failure in any of the calls below still reaches the
	// Terraform state. Dropping it would orphan the service and make every later apply
	// fail with "a helm named X already exists".
	partial, partialErr := newDomainHelmFromQovery(newHelm, request.DeploymentStageID, request.IsSkipped, request.AdvancedSettingsJson, nil)
	if partialErr != nil {
		// The response cannot be represented as a domain helm service, but it does exist.
		// Fall back to its identifiers: writing the ID is the whole point here.
		partial = identityOnlyHelm(newHelm, environmentID)
	}

	// Create custom domains
	if !request.CustomDomains.IsEmpty() {
		for _, customDomain := range request.CustomDomains.Create {
			_, resp, err := c.client.HelmCustomDomainAPI.
				CreateHelmCustomDomain(ctx, newHelm.Id).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return partial, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, request.Name, resp, err)
			}
		}
	}

	// Attach helm to deployment stage
	if len(request.DeploymentStageID) > 0 {
		response, err := attachServiceToDeploymentStage(ctx, c.client, request.DeploymentStageID, newHelm.Id, request.IsSkipped)
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return partial, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.HELM, newHelm.Id, request.AdvancedSettingsJson)
	if err != nil {
		return partial, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, nil, err)
	}

	// Get helm deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, newHelm.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return partial, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, newHelm.Id, resp, err)
	}

	// Get custom domains
	customDomains, resp, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, newHelm.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return partial, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, newHelm.Id, resp, err)
	}

	newHelmDomain, err := newDomainHelmFromQovery(newHelm, deploymentStage.Id, getServiceIsSkipped(deploymentStage, newHelm.Id), request.AdvancedSettingsJson, customDomains)
	if err != nil {
		// Every call succeeded but the response still cannot be converted. The service
		// exists, so hand back the fallback rather than losing it to the state.
		return partial, err
	}

	return newHelmDomain, nil
}

// Get calls Qovery's API to retrieve a helm using the given helmID.
func (c helmQoveryAPI) Get(ctx context.Context, helmID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*helm.Helm, error) {
	helm, resp, err := c.client.HelmMainCallsAPI.
		GetHelm(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	// Get helm deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, helmID).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).ReadServiceAdvancedSettings(domain.HELM, helmID, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelm, helmID, nil, err)
	}

	// Get custom domains
	customDomains, resp, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, helm.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelmCustomDomain, helm.Id, resp, err)
	}

	return newDomainHelmFromQovery(helm, deploymentStage.Id, getServiceIsSkipped(deploymentStage, helm.Id), *advancedSettingsAsJson, customDomains)
}

// Update calls Qovery's API to update a helm using the given helmID and request.
func (c helmQoveryAPI) Update(ctx context.Context, helmID string, request helm.UpsertRepositoryRequest) (*helm.Helm, error) {
	req, err := newQoveryHelmRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrInvalidHelmUpsertRequest.Error())
	}

	helm, resp, err := c.client.HelmMainCallsAPI.
		EditHelm(ctx, helmID).
		HelmRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	// Create custom domains
	if !request.CustomDomains.IsEmpty() {
		for _, customDomain := range request.CustomDomains.Delete {
			_, err := c.client.HelmCustomDomainAPI.
				DeleteHelmCustomDomain(ctx, helmID, customDomain.Id).Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, request.Name, resp, err)
			}
		}
		for _, customDomain := range request.CustomDomains.Update {
			_, resp, err := c.client.HelmCustomDomainAPI.
				EditHelmCustomDomain(ctx, helmID, customDomain.Id).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, request.Name, resp, err)
			}
		}
		for _, customDomain := range request.CustomDomains.Create {
			_, resp, err := c.client.HelmCustomDomainAPI.
				CreateHelmCustomDomain(ctx, helmID).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, request.Name, resp, err)
			}
		}
	}

	// Attach helm to deployment stage
	if len(request.DeploymentStageID) > 0 {
		response, err := attachServiceToDeploymentStage(ctx, c.client, request.DeploymentStageID, helmID, request.IsSkipped)
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelm, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.HELM, helmID, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, nil, err)
	}

	// Get helm deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, helmID).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	// Get custom domains
	customDomains, resp, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, helm.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelmCustomDomain, helm.Id, resp, err)
	}

	return newDomainHelmFromQovery(helm, deploymentStage.Id, getServiceIsSkipped(deploymentStage, helm.Id), request.AdvancedSettingsJson, customDomains)
}

// Delete calls Qovery's API to deletes a helm using the given helmID.
func (c helmQoveryAPI) Delete(ctx context.Context, helmID string) error {
	_, resp, err := c.client.HelmMainCallsAPI.
		GetHelm(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// if the helm is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	resp, err = c.client.HelmMainCallsAPI.
		DeleteHelm(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	return nil
}

// identityOnlyHelm is the last resort when a freshly created helm service cannot be
// converted from its API response. Only the identifiers matter: the resource layer needs
// the ID in the Terraform state so the service gets tainted and replaced rather than
// orphaned. Returns nil when even the identifiers make no sense.
func identityOnlyHelm(h *qovery.HelmResponse, requestedEnvironmentID string) *helm.Helm {
	if h == nil {
		return nil
	}

	helmID, err := uuid.Parse(h.Id)
	if err != nil {
		return nil
	}

	// Prefer the environment the response reports, but fall back to the one the create was
	// aimed at: a response missing or mangling its environment reference must not cost us
	// the whole entity.
	environmentID, err := uuid.Parse(h.Environment.Id)
	if err != nil {
		environmentID, err = uuid.Parse(requestedEnvironmentID)
		if err != nil {
			return nil
		}
	}

	return &helm.Helm{
		ID:            helmID,
		EnvironmentID: environmentID,
		Name:          h.Name,
	}
}
