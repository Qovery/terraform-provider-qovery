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

func (c DeploymentDesiredState) String() string {
	switch c {
	case RUNNING:
		return "RUNNING"
	case STOPPED:
		return "STOPPED"
	case RESTARTED:
		return "RESTARTED"
	case DELETED:
		return "DELETED"
	}

	return "UNDEFINED"
}

type Deployment struct {
	EnvironmentId *uuid.UUID
	DesiredState  DeploymentDesiredState
	ForceTrigger  string
}

type NewDeploymentParams struct {
	EnvironmentId string
	DesiredState  string
	ForceTrigger  string
}

func NewDeployment(params NewDeploymentParams) (*Deployment, error) {
	// Check desired state
	desiredState, err := fromString(params.DesiredState)
	if err != nil {
		return nil, err
	}

	environmentUuid, err := uuid.Parse(params.EnvironmentId)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIdParam.Error())
	}

	return &Deployment{
		EnvironmentId: &environmentUuid,
		DesiredState:  *desiredState,
		ForceTrigger:  params.ForceTrigger,
	}, nil
}
