package qoveryapi

import (
	"context"

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
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, request.Name, resp, err)
			}
		}
	}

	// Attach helm to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, newHelm.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, resp, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.HELM, newHelm.Id, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, nil, err)
	}

	// Get helm deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, newHelm.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, newHelm.Id, resp, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, newHelm.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, newHelm.Id, resp, err)
	}

	return newDomainHelmFromQovery(newHelm, deploymentStage.Id, request.AdvancedSettingsJson, customDomains)
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
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).ReadServiceAdvancedSettings(domain.HELM, helmID, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelm, helmID, nil, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, helm.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, helm.Id, resp, err)
	}

	return newDomainHelmFromQovery(helm, deploymentStage.Id, *advancedSettingsAsJson, customDomains)
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
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, helmID).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, resp, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.HELM, helmID, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, request.Name, nil, err)
	}

	// Get helm deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, helmID).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.HelmCustomDomainAPI.ListHelmCustomDomain(ctx, helm.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmCustomDomain, helm.Id, resp, err)
	}

	return newDomainHelmFromQovery(helm, deploymentStage.Id, request.AdvancedSettingsJson, customDomains)
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
