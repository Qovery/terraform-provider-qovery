package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// Ensure blueprintService defined type fully satisfy the blueprint.Service interface.
var _ blueprint.Service = blueprintService{}

// blueprintService implements the interface blueprint.Service.
type blueprintService struct {
	blueprintRepository blueprint.Repository
	waitTimeout         time.Duration
	waitPollInterval    time.Duration
}

func NewBlueprintService(blueprintRepository blueprint.Repository) (blueprint.Service, error) {
	if blueprintRepository == nil {
		return nil, ErrInvalidRepository
	}
	return &blueprintService{
		blueprintRepository: blueprintRepository,
		waitTimeout:         defaultWaitTimeout,
		waitPollInterval:    defaultWaitPollInterval,
	}, nil
}

func (s blueprintService) Create(ctx context.Context, environmentID string, request blueprint.CreateRequest) (*blueprint.Blueprint, error) {
	if err := validateUUIDParam(environmentID, blueprint.ErrInvalidEnvironmentIDParam); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToCreateBlueprint.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToCreateBlueprint.Error())
	}

	created, err := s.blueprintRepository.Create(ctx, environmentID, request)
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToCreateBlueprint.Error())
	}

	// The blueprint exists from here on: every failure below returns it so the resource
	// layer can persist its ID instead of leaving an untracked blueprint in Qovery.
	bp, err := s.converge(ctx, created.BlueprintID, created.DispatchID, request.Deploy)
	if err != nil {
		if bp == nil {
			bp = newBlueprintStub(created.BlueprintID, environmentID, request.UpsertRequest)
		}
		return bp, errors.Wrap(err, blueprint.ErrFailedToCreateBlueprint.Error())
	}
	return bp, nil
}

// newBlueprintStub stands in when no read succeeded after create, so state still gets the ID
func newBlueprintStub(blueprintID string, environmentID string, request blueprint.UpsertRequest) *blueprint.Blueprint {
	id, err := uuid.Parse(blueprintID)
	if err != nil {
		return nil
	}
	return &blueprint.Blueprint{
		ID:            id,
		EnvironmentID: uuid.MustParse(environmentID),
		Name:          request.Name,
		Tag:           request.Tag,
	}
}

func (s blueprintService) Get(ctx context.Context, blueprintID string) (*blueprint.Blueprint, error) {
	if err := validateUUIDParam(blueprintID, blueprint.ErrInvalidBlueprintIDParam); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToGetBlueprint.Error())
	}
	bp, err := s.blueprintRepository.Get(ctx, blueprintID)
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToGetBlueprint.Error())
	}
	if bp.ServiceID != nil {
		bp.ServiceStatus, err = s.blueprintRepository.GetServiceStatus(ctx, bp.EnvironmentID.String(), bp.ServiceType, *bp.ServiceID)
		if err != nil {
			return nil, errors.Wrap(err, blueprint.ErrFailedToGetBlueprint.Error())
		}
	}
	return bp, nil
}

func (s blueprintService) ResolveLatestTag(ctx context.Context, environmentID string, version blueprint.CatalogVersion) (string, error) {
	if err := validateUUIDParam(environmentID, blueprint.ErrInvalidEnvironmentIDParam); err != nil {
		return "", errors.Wrap(err, blueprint.ErrFailedToResolveTag.Error())
	}
	organizationID, err := s.blueprintRepository.GetOrganizationID(ctx, environmentID)
	if err != nil {
		return "", errors.Wrap(err, blueprint.ErrFailedToResolveTag.Error())
	}
	entries, err := s.blueprintRepository.ListCatalog(ctx, organizationID)
	if err != nil {
		return "", errors.Wrap(err, blueprint.ErrFailedToResolveTag.Error())
	}

	var available []string
	for _, entry := range entries {
		if entry.Equal(version) {
			return entry.LatestTag, nil
		}
		if entry.SameService(version) {
			available = append(available, entry.String())
		}
	}
	if len(available) > 0 {
		return "", errors.Wrapf(blueprint.ErrCatalogVersionNotFound, "%s, available: %s", version, strings.Join(available, ", "))
	}
	return "", errors.Wrap(blueprint.ErrCatalogVersionNotFound, version.String())
}

func (s blueprintService) Update(ctx context.Context, blueprintID string, request blueprint.UpdateRequest) (*blueprint.Blueprint, error) {
	if err := validateUUIDParam(blueprintID, blueprint.ErrInvalidBlueprintIDParam); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToUpdateBlueprint.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToUpdateBlueprint.Error())
	}

	if err := s.blueprintRepository.Update(ctx, blueprintID, request); err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToUpdateBlueprint.Error())
	}
	dispatchID, err := s.blueprintRepository.Deploy(ctx, blueprintID)
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToUpdateBlueprint.Error())
	}

	bp, err := s.converge(ctx, blueprintID, dispatchID, true)
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrFailedToUpdateBlueprint.Error())
	}
	return bp, nil
}

func (s blueprintService) Delete(ctx context.Context, blueprintID string) error {
	if err := validateUUIDParam(blueprintID, blueprint.ErrInvalidBlueprintIDParam); err != nil {
		return errors.Wrap(err, blueprint.ErrFailedToDeleteBlueprint.Error())
	}

	bp, err := s.blueprintRepository.Get(ctx, blueprintID)
	if err != nil {
		if apierrors.IsErrNotFound(err) {
			return nil
		}
		return errors.Wrap(err, blueprint.ErrFailedToDeleteBlueprint.Error())
	}
	if bp.ServiceID == nil {
		// No service to delete, and the API has no blueprint delete endpoint: the row stays.
		tflog.Warn(ctx, fmt.Sprintf("blueprint %s has no linked service, it cannot be deleted through the API and is only removed from the Terraform state", blueprintID))
		return nil
	}

	if err := s.blueprintRepository.DeleteService(ctx, bp.ServiceType, *bp.ServiceID); err != nil {
		return errors.Wrap(err, blueprint.ErrFailedToDeleteBlueprint.Error())
	}

	// Qovery deletes the blueprint once the engine has deleted its service.
	subject := fmt.Sprintf("blueprint %s to be deleted", blueprintID)
	if err := s.wait(ctx, s.blueprintDeletedFunc(blueprintID), subject); err != nil {
		return errors.Wrap(err, blueprint.ErrFailedToDeleteBlueprint.Error())
	}
	return nil
}

// converge waits for the dispatch to finish then, when deploying, for the service
// deployment it chains. It returns the blueprint as last read, even alongside an error.
func (s blueprintService) converge(ctx context.Context, blueprintID string, dispatchID string, deploy bool) (*blueprint.Blueprint, error) {
	var last *blueprint.Blueprint
	dispatchSubject := fmt.Sprintf("blueprint %s dispatch %s to finish", blueprintID, dispatchID)
	err := s.wait(ctx, func(ctx context.Context) (bool, error) {
		bp, err := s.blueprintRepository.Get(ctx, blueprintID)
		if err != nil {
			return false, err
		}
		last = bp
		return dispatchFinished(bp, dispatchID)
	}, dispatchSubject)
	if err != nil {
		return last, err
	}
	if last.ServiceID == nil {
		return last, errors.Wrap(blueprint.ErrDispatchDidNotCreateService, blueprintID)
	}
	if !deploy {
		return last, nil
	}

	deploySubject := fmt.Sprintf("service %s of blueprint %s to be deployed", *last.ServiceID, blueprintID)
	err = s.wait(ctx, s.serviceDeployedFunc(last), deploySubject)
	return last, err
}

// wait reports a timeout as blueprint.ErrWaitTimeout, not the deployment one the shared helper uses
func (s blueprintService) wait(ctx context.Context, f waitFunc, subject string) error {
	err := wait(ctx, f, subject, s.waitTimeout, s.waitPollInterval)
	if errors.Is(err, deployment.ErrWaitTimeout) {
		return fmt.Errorf("%w: waited %s for %s", blueprint.ErrWaitTimeout, s.waitTimeout, subject)
	}
	return err
}

// dispatchFinished reports whether the dispatch dispatchID ended. latest_deployment can
// still be an earlier dispatch right after the API call, which counts as not finished.
func dispatchFinished(bp *blueprint.Blueprint, dispatchID string) (bool, error) {
	d := bp.LatestDeployment
	if d == nil || d.ID != dispatchID {
		return false, nil
	}
	switch {
	case d.Status.IsPending():
		return false, nil
	case d.Status.IsSuccess():
		return true, nil
	case d.Status.IsFailure():
		message := string(d.Status)
		if d.ErrorMessage != nil {
			message = fmt.Sprintf("%s: %s", d.Status, *d.ErrorMessage)
		}
		return false, errors.Wrap(blueprint.ErrDispatchFailed, message)
	}
	return false, errors.Wrap(blueprint.ErrUnexpectedDispatchStatus, string(d.Status))
}

// Service deploy is chained after the dispatch: only a deployment newer than its start is ours
func (s blueprintService) serviceDeployedFunc(bp *blueprint.Blueprint) waitFunc {
	dispatchStartedAt := bp.LatestDeployment.StartedAt
	return func(ctx context.Context) (bool, error) {
		serviceStatus, err := s.blueprintRepository.GetServiceStatus(ctx, bp.EnvironmentID.String(), bp.ServiceType, *bp.ServiceID)
		if err != nil {
			return false, err
		}
		// Callers read LastApplyFailed off the returned blueprint, so it must see this status
		bp.ServiceStatus = serviceStatus
		if serviceStatus == nil || serviceStatus.LastDeploymentDate == nil || !serviceStatus.LastDeploymentDate.After(dispatchStartedAt) {
			return false, nil
		}
		st := status.Status{State: status.State(serviceStatus.State)}
		// Bare QUEUED has no _QUEUED suffix, so IsFinalState counts it final
		if !st.IsFinalState() || st.State == status.StateQueued {
			return false, nil
		}
		if st.State != status.StateDeployed {
			return false, errors.Wrap(blueprint.ErrServiceDeploymentFailed, serviceStatus.State)
		}
		return true, nil
	}
}

func (s blueprintService) blueprintDeletedFunc(blueprintID string) waitFunc {
	return func(ctx context.Context) (bool, error) {
		_, err := s.blueprintRepository.Get(ctx, blueprintID)
		if err == nil {
			return false, nil
		}
		if apierrors.IsErrNotFound(err) {
			return true, nil
		}
		return false, err
	}
}
