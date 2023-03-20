package deploymentstage

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidDeploymentStageIDParam is returned if the deployment stage ID indicated is not valid
	ErrInvalidDeploymentStageIDParam = errors.New("invalid deployment stage ID")
	// ErrInvalidEnvironmentIDParam is returned if the environment ID indicated is not valid
	ErrInvalidEnvironmentIDParam = errors.New("invalid environment ID")
	// ErrInvalidDeploymentStageNameParam is returned if the deployment stage name indicated is not valid
	ErrInvalidDeploymentStageNameParam = errors.New("invalid deployment stage name")
	// ErrInvalidDeploymentStage is returned if the validation fails
	ErrInvalidDeploymentStage = errors.New("invalid deployment stage")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid deployment stage upsert request")
	// ErrInvalidMoveAfterParam is returned if the move_after ID indicated is not valid
	ErrInvalidMoveAfterParam = errors.New("invalid move_after param")
	// ErrInvalidMoveBeforeParam is returned if the move_before ID indicated is not valid
	ErrInvalidMoveBeforeParam = errors.New("invalid move_before param")
)

type DeploymentStage struct {
	ID            uuid.UUID
	EnvironmentID uuid.UUID
	Name          string
	Description   string
	MoveAfter     *uuid.UUID
	MoveBefore    *uuid.UUID
}

// NewDeploymentStageParams represents the arguments needed to create a DeploymentStage.
type NewDeploymentStageParams struct {
	DeploymentStageID string
	EnvironmentID     string
	Name              string
	Description       string
	MoveAfter         *string
	MoveBefore        *string
}

// Validate returns an error to tell whether the DeploymentStage domain model is valid or not.
func (p DeploymentStage) Validate() error {
	return validator.New().Struct(p)
}

// NewDeploymentStage returns a new instance of a DeploymentStage domain model.
func NewDeploymentStage(params NewDeploymentStageParams) (*DeploymentStage, error) {
	deploymentStageUuid, err := uuid.Parse(params.DeploymentStageID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidDeploymentStageIDParam.Error())
	}

	environmentUuid, err := uuid.Parse(params.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIDParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidDeploymentStageNameParam
	}

	var moveAfter *uuid.UUID = nil
	if params.MoveAfter != nil {
		newMoveAfter, err := uuid.Parse(*params.MoveAfter)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidMoveAfterParam.Error())
		}
		moveAfter = &newMoveAfter
	}

	var moveBefore *uuid.UUID = nil
	if params.MoveBefore != nil {
		newMoveBefore, err := uuid.Parse(*params.MoveBefore)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidMoveBeforeParam.Error())
		}
		moveBefore = &newMoveBefore
	}

	v := &DeploymentStage{
		ID:            deploymentStageUuid,
		EnvironmentID: environmentUuid,
		Name:          params.Name,
		Description:   params.Description,
		MoveAfter:     moveAfter,
		MoveBefore:    moveBefore,
	}

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidDeploymentStage.Error())
	}

	return v, nil
}
