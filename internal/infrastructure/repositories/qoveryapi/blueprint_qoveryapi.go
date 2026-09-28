package qoveryapi

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

// Ensure blueprintQoveryAPI defined type fully satisfy the blueprint.Repository interface.
var _ blueprint.Repository = blueprintQoveryAPI{}

// blueprintQoveryAPI implements the interface blueprint.Repository.
type blueprintQoveryAPI struct {
	client *qovery.APIClient
}

func newBlueprintQoveryAPI(client *qovery.APIClient) (blueprint.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &blueprintQoveryAPI{client: client}, nil
}

func (c blueprintQoveryAPI) Create(ctx context.Context, environmentID string, request blueprint.CreateRequest) (*blueprint.CreationResult, error) {
	created, resp, err := c.client.BlueprintMainCallsAPI.
		CreateBlueprintDeployment(ctx, environmentID).
		Deploy(request.Deploy).
		BlueprintCreateRequest(newQoveryBlueprintCreateRequest(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceBlueprint, request.Name, resp, err)
	}
	return &blueprint.CreationResult{
		BlueprintID: created.GetId(),
		DispatchID:  created.GetDeploymentId(),
	}, nil
}

func (c blueprintQoveryAPI) Get(ctx context.Context, blueprintID string) (*blueprint.Blueprint, error) {
	details, resp, err := c.client.BlueprintMainCallsAPI.
		GetBlueprint(ctx, blueprintID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceBlueprint, blueprintID, resp, err)
	}

	variables, resp, err := c.client.BlueprintMainCallsAPI.
		GetBlueprintVariables(ctx, blueprintID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceBlueprint, blueprintID, resp, err)
	}

	return newDomainBlueprintFromQovery(details, variables)
}

func (c blueprintQoveryAPI) Update(ctx context.Context, blueprintID string, request blueprint.UpdateRequest) error {
	_, resp, err := c.client.BlueprintMainCallsAPI.
		UpdateBlueprint(ctx, blueprintID).
		BlueprintUpdateRequest(newQoveryBlueprintUpdateRequest(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return apierrors.NewUpdateAPIError(apierrors.APIResourceBlueprint, blueprintID, resp, err)
	}
	return nil
}

func (c blueprintQoveryAPI) Deploy(ctx context.Context, blueprintID string) (string, error) {
	ack, resp, err := c.client.BlueprintMainCallsAPI.
		DeployBlueprint(ctx, blueprintID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return "", apierrors.NewDeployAPIError(apierrors.APIResourceBlueprint, blueprintID, resp, err)
	}
	return ack.GetDeploymentId(), nil
}

func (c blueprintQoveryAPI) GetServiceStatus(ctx context.Context, environmentID string, serviceType blueprint.ServiceType, serviceID string) (*blueprint.ServiceStatus, error) {
	statuses, resp, err := c.client.EnvironmentMainCallsAPI.
		GetEnvironmentStatuses(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceEnvironmentStatus, environmentID, resp, err)
	}
	return findServiceStatus(statuses, serviceType, serviceID)
}

func (c blueprintQoveryAPI) GetServiceIconURI(ctx context.Context, serviceType blueprint.ServiceType, serviceID string) (*string, error) {
	var (
		iconURI  string
		resp     *http.Response
		err      error
		resource apierrors.APIResource
	)
	switch serviceType {
	case blueprint.ServiceTypeTerraform:
		var service *qovery.TerraformResponse
		resource = apierrors.APIResourceTerraformService
		service, resp, err = c.client.TerraformMainCallsAPI.GetTerraform(ctx, serviceID).Execute()
		iconURI = service.GetIconUri()
	case blueprint.ServiceTypeHelm:
		var service *qovery.HelmResponse
		resource = apierrors.APIResourceHelm
		service, resp, err = c.client.HelmMainCallsAPI.GetHelm(ctx, serviceID).Execute()
		iconURI = service.GetIconUri()
	default:
		return nil, errors.Wrap(blueprint.ErrUnknownServiceType, string(serviceType))
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(resource, serviceID, resp, err)
	}
	return &iconURI, nil
}

func (c blueprintQoveryAPI) GetOrganizationID(ctx context.Context, environmentID string) (string, error) {
	environment, resp, err := c.client.EnvironmentMainCallsAPI.
		GetEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return "", apierrors.NewReadAPIError(apierrors.APIResourceEnvironment, environmentID, resp, err)
	}
	return environment.Organization.GetId(), nil
}

func (c blueprintQoveryAPI) ListCatalog(ctx context.Context, organizationID string) ([]blueprint.CatalogEntry, error) {
	catalog, resp, err := c.client.BlueprintMainCallsAPI.
		GetBlueprintCatalog(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceBlueprint, "catalog", resp, err)
	}
	return newDomainCatalogEntriesFromQovery(catalog), nil
}

func (c blueprintQoveryAPI) GetVariableDefaults(ctx context.Context, organizationID string, environmentID string, version blueprint.CatalogVersion) (map[string]string, error) {
	manifest, resp, err := c.client.BlueprintCatalogAPI.
		GetBlueprintCatalogServiceManifest(ctx, organizationID, version.Provider, version.ServiceFamily, version.ServiceVersion).
		EnvironmentId(environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceBlueprint, version.String(), resp, err)
	}
	return newVariableDefaultsFromQovery(manifest), nil
}

func (c blueprintQoveryAPI) DeleteService(ctx context.Context, serviceType blueprint.ServiceType, serviceID string) (bool, error) {
	var (
		resp     *http.Response
		err      error
		resource apierrors.APIResource
	)
	switch serviceType {
	case blueprint.ServiceTypeTerraform:
		resource = apierrors.APIResourceTerraformService
		resp, err = c.client.TerraformMainCallsAPI.DeleteTerraform(ctx, serviceID).Execute()
	case blueprint.ServiceTypeHelm:
		resource = apierrors.APIResourceHelm
		resp, err = c.client.HelmMainCallsAPI.DeleteHelm(ctx, serviceID).Execute()
	default:
		return false, errors.Wrap(blueprint.ErrUnknownServiceType, string(serviceType))
	}
	// Only 404: a 403 can be a real permission error, and "gone" skips the deletion wait
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err != nil || resp.StatusCode >= 300 {
		return false, apierrors.NewDeleteAPIError(resource, serviceID, resp, err)
	}
	return true, nil
}
