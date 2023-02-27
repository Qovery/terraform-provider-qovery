package newdeployment

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidEnvironmentIdParam is returned if the environment id indicated is not valid
	ErrInvalidEnvironmentIdParam = errors.New("invalid environment Id")
	// ErrInvalidServiceIdParam is returned if the environment id indicated is not valid
	ErrInvalidServiceIdParam = errors.New("invalid service Id")
	// ErrInvalidDeployment is returned if deployment is incoherent
	ErrInvalidDeployment = errors.New("invalid deployment")
	// ErrInvalidDeploymentDesiredState is returned if the deployment desired state is incoherent
	ErrInvalidDeploymentDesiredState = errors.New("invalid deployment desired state")
)

type DeploymentDesiredState string

const (
	RUNNING   DeploymentDesiredState = "RUNNING"
	STOPPED   DeploymentDesiredState = "STOPPED"
	RESTARTED DeploymentDesiredState = "RESTARTED"
	DELETED   DeploymentDesiredState = "DELETED"
)

func fromString(desiredStateStr string) (*DeploymentDesiredState, error) {
	desiredState := DeploymentDesiredState(desiredStateStr)
	switch desiredState {
	case RUNNING, STOPPED, RESTARTED, DELETED:
		return &desiredState, nil
	}
	return nil, ErrInvalidDeploymentDesiredState
}

type Deployment struct {
	EnvironmentId *uuid.UUID
	ServiceIds    []uuid.UUID
	DesiredState  DeploymentDesiredState
}

func (d Deployment) HasServiceIds() bool {
	if d.ServiceIds == nil {
		return false
	}
	return len(d.ServiceIds) > 0
}

type NewDeploymentParams struct {
	EnvironmentId string
	ServiceIds    []string
	DesiredState  string
}

func NewDeployment(params NewDeploymentParams) (*Deployment, error) {
	serviceIdsIsDefined := len(params.ServiceIds) > 0

	// Check desired state
	desiredState, err := fromString(params.DesiredState)
	if err != nil {
		return nil, err
	}

	// If environment id is defined, then validate uuid and create Environment DeploymentEnvironment
	environmentUuid, err := uuid.Parse(params.EnvironmentId)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIdParam.Error())
	}

	// If service ids are defined, then validate uuids
	var serviceUuids []uuid.UUID = nil
	if serviceIdsIsDefined {
		serviceUuids = make([]uuid.UUID, 0, len(params.ServiceIds))
		for _, serviceId := range params.ServiceIds {
			serviceUuid, err := uuid.Parse(serviceId)
			if err != nil {
				return nil, errors.Wrap(err, ErrInvalidServiceIdParam.Error())
			}
			serviceUuids = append(serviceUuids, serviceUuid)
		}
	}

	return &Deployment{
		EnvironmentId: &environmentUuid,
		ServiceIds:    serviceUuids,
		DesiredState:  *desiredState,
	}, nil
}
