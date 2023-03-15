package newdeployment

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidIdParam is returned if the id indicated is not valid
	ErrInvalidIdParam = errors.New("invalid Id")
	// ErrInvalidVersionParam is returned if the version indicated is not valid
	ErrInvalidVersionParam = errors.New("invalid Version")
	// ErrInvalidEnvironmentIdParam is returned if the environment id indicated is not valid
	ErrInvalidEnvironmentIdParam = errors.New("invalid environment Id")
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
	Id            *uuid.UUID
	EnvironmentId *uuid.UUID
	Version       *uuid.UUID
	DesiredState  DeploymentDesiredState
}

type NewDeploymentParams struct {
	Id            *string
	EnvironmentId string
	Version       *string
	DesiredState  string
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

	var id uuid.UUID
	// If unset, generate a random one
	if params.Id == nil {
		id = uuid.New()
	} else {
		id, err = uuid.Parse(*params.Id)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidIdParam.Error())
		}
	}

	var version *uuid.UUID = nil
	if params.Version != nil {
		newVersion, err := uuid.Parse(*params.Version)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidVersionParam.Error())
		}
		version = &newVersion
	}

	return &Deployment{
		Id:            &id,
		EnvironmentId: &environmentUuid,
		Version:       version,
		DesiredState:  *desiredState,
	}, nil
}
