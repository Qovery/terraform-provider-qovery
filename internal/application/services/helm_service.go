package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure helmService defined types fully satisfy the helm.Service interface.
var _ helm.Service = helmService{}

// helmService implements the interface helm.Service.
type helmService struct {
	helmRepository               helm.Repository
	helmDeploymentService        deployment.Service
	variableService              variable.Service
	secretService                secret.Service
	deploymentRestrictionService deploymentrestriction.DeploymentRestrictionService
}

// NewHelmService return a new instance of a helm.Service that uses the given helm.Repository.
func NewHelmService(
	helmRepository helm.Repository,
	helmDeploymentService deployment.Service,
	variableService variable.Service,
	secretService secret.Service,
	deploymentRestrictionService deploymentrestriction.DeploymentRestrictionService,
) (helm.Service, error) {
	if helmRepository == nil {
		return nil, ErrInvalidRepository
	}

	if helmDeploymentService == nil {
		return nil, ErrInvalidService
	}

	if variableService == nil {
		return nil, ErrInvalidService
	}

	if secretService == nil {
		return nil, ErrInvalidService
	}

	return &helmService{
		helmRepository:               helmRepository,
		variableService:              variableService,
		secretService:                secretService,
		helmDeploymentService:        helmDeploymentService,
		deploymentRestrictionService: deploymentRestrictionService,
	}, nil
}

// Create handles the domain logic to create a helm.
func (s helmService) Create(ctx context.Context, environmentID string, request helm.UpsertServiceRequest) (*helm.Helm, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	newHelm, err := s.helmRepository.Create(ctx, environmentID, request.HelmUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	overridesAuthorizedScopes := make(map[variable.Scope]struct{})
	overridesAuthorizedScopes[variable.ScopeProject] = struct{}{}
	overridesAuthorizedScopes[variable.ScopeEnvironment] = struct{}{}
	_, err = s.variableService.Update(ctx, newHelm.ID.String(), request.EnvironmentVariables, request.EnvironmentVariableAliases, request.EnvironmentVariableOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	_, err = s.secretService.Update(ctx, newHelm.ID.String(), request.Secrets, request.SecretAliases, request.SecretOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	if request.DeploymentRestrictionsDiff.IsNotEmpty() {
		if apiErr := s.deploymentRestrictionService.UpdateServiceDeploymentRestrictions(ctx, newHelm.ID.String(), domain.HELM, request.DeploymentRestrictionsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	newHelm, err = s.refreshHelm(ctx, *newHelm)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToCreateHelm.Error())
	}

	return newHelm, nil
}

// Get handles the domain logic to retrieve a helm.
func (s helmService) Get(ctx context.Context, helmID string, advancedSettingsJsonFromState string) (*helm.Helm, error) {
	if err := s.checkID(helmID); err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToGetHelm.Error())
	}

	fetchedHelm, err := s.helmRepository.Get(ctx, helmID, advancedSettingsJsonFromState)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToGetHelm.Error())
	}

	fetchedHelm, err = s.refreshHelm(ctx, *fetchedHelm)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToGetHelm.Error())
	}

	return fetchedHelm, nil
}

// Update handles the domain logic to update a helm.
func (s helmService) Update(ctx context.Context, helmID string, request helm.UpsertServiceRequest) (*helm.Helm, error) {
	if err := s.checkID(helmID); err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	updateHelm, err := s.helmRepository.Update(ctx, helmID, request.HelmUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	overridesAuthorizedScopes := make(map[variable.Scope]struct{})
	overridesAuthorizedScopes[variable.ScopeProject] = struct{}{}
	overridesAuthorizedScopes[variable.ScopeEnvironment] = struct{}{}
	_, err = s.variableService.Update(ctx, updateHelm.ID.String(), request.EnvironmentVariables, request.EnvironmentVariableAliases, request.EnvironmentVariableOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	_, err = s.secretService.Update(ctx, updateHelm.ID.String(), request.Secrets, request.SecretAliases, request.SecretOverrides, overridesAuthorizedScopes)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	if request.DeploymentRestrictionsDiff.IsNotEmpty() {
		if apiErr := s.deploymentRestrictionService.UpdateServiceDeploymentRestrictions(ctx, updateHelm.ID.String(), domain.HELM, request.DeploymentRestrictionsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	updateHelm, err = s.refreshHelm(ctx, *updateHelm)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrFailedToUpdateHelm.Error())
	}

	return updateHelm, nil
}

// Delete handles the domain logic to delete a helm.
func (s helmService) Delete(ctx context.Context, helmID string) error {
	if err := s.checkID(helmID); err != nil {
		return errors.Wrap(err, helm.ErrFailedToDeleteHelm.Error())
	}

	if err := s.helmRepository.Delete(ctx, helmID); err != nil {
		return errors.Wrap(err, helm.ErrFailedToDeleteHelm.Error())
	}

	if err := wait(ctx, waitNotFoundFunc(s.helmDeploymentService, helmID)); err != nil {
		return errors.Wrap(err, helm.ErrFailedToDeleteHelm.Error())
	}

	return nil
}

func (s helmService) refreshHelm(ctx context.Context, helm helm.Helm) (*helm.Helm, error) {
	envVars, err := s.variableService.List(ctx, helm.ID.String())
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretService.List(ctx, helm.ID.String())
	if err != nil {
		return nil, err
	}

	status, err := s.helmDeploymentService.GetStatus(ctx, helm.ID.String())
	if err != nil {
		return nil, err
	}

	deploymentRestrictions, apiErr := s.deploymentRestrictionService.GetServiceDeploymentRestrictions(ctx, helm.ID.String(), domain.HELM)
	if apiErr != nil {
		return nil, apiErr
	}

	if err := helm.SetEnvironmentVariables(envVars); err != nil {
		return nil, err
	}

	if err := helm.SetSecrets(secrets); err != nil {
		return nil, err
	}

	if err := helm.SetState(status.State); err != nil {
		return nil, err
	}

	helm.JobDeploymentRestrictions = deploymentRestrictions

	return &helm, err
}

// checkEnvironmentID validates that the given environmentID is valid.
func (s helmService) checkEnvironmentID(environmentID string) error {
	if environmentID == "" {
		return helm.ErrInvalidHelmEnvironmentIDParam
	}

	if _, err := uuid.Parse(environmentID); err != nil {
		return errors.Wrap(err, helm.ErrInvalidHelmEnvironmentIDParam.Error())
	}

	return nil
}

// checkID validates that the given helmID is valid.
func (s helmService) checkID(helmID string) error {
	if helmID == "" {
		return helm.ErrInvalidHelmIDParam
	}

	if _, err := uuid.Parse(helmID); err != nil {
		return errors.Wrap(err, helm.ErrInvalidHelmIDParam.Error())
	}

	return nil
}
