package status

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilStatus is returned if a Status is nil.
	ErrNilStatus = errors.New("status cannot be nil")
	// ErrInvalidStatus is the error return if a Status is invalid.
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidStatusIDParam is returned if the status id param is invalid.
	ErrInvalidStatusIDParam = errors.New("invalid status id param")
	// ErrInvalidStateParam is returned if the state param is invalid.
	ErrInvalidStateParam = errors.New("invalid state param")
	// ErrInvalidServiceDeploymentStatusParam is returned if the service deployment status param is invalid.
	ErrInvalidServiceDeploymentStatusParam = errors.New("invalid service deployment status param")
	// ErrInvalidLastDeploymentDateParam is returned if the last deployment date param is invalid.
	ErrInvalidLastDeploymentDateParam = errors.New("invalid last deployment date param")
)

type Status struct {
	ID                 uuid.UUID `validate:"required"`
	State              State     `validate:"required"`
	LastDeploymentDate *time.Time
}

// Validate returns an error to tell whether the Status domain model is valid or not.
func (s Status) Validate() error {
	if err := validator.New().Struct(s); err != nil {
		return errors.Wrap(err, ErrInvalidStatus.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Status domain model is valid or not.
func (s Status) IsValid() bool {
	return s.Validate() == nil
}

// IsFinalState returns a bool to tell whether the Status is a final state or not.
func (s Status) IsFinalState() bool {
	return !s.IsProcessingState() &&
		!s.IsWaitingState() &&
		!s.IsQueuedState()
}

// IsErrorState returns a bool to tell whether the Status is an error state or not.
func (s Status) IsErrorState() bool {
	return strings.HasSuffix(s.State.String(), "_ERROR")
}

// IsWaitingState returns a bool to tell whether the Status is a waiting state or not.
func (s Status) IsWaitingState() bool {
	return strings.HasSuffix(s.State.String(), "_WAITING")
}

// IsQueuedState returns a bool to tell whether the Status is a queued state or not.
func (s Status) IsQueuedState() bool {
	return strings.HasSuffix(s.State.String(), "_QUEUED")
}

// IsProcessingState returns a bool to tell whether the Status is a processing state or not.
func (s Status) IsProcessingState() bool {
	return strings.HasSuffix(s.State.String(), "ING")
}

// NewStatusParams represents the arguments needed to create a Status.
type NewStatusParams struct {
	StatusID                string
	State                   string
	ServiceDeploymentStatus string
	LastDeploymentDate      *time.Time
}

// NewEnvironmentStatusParams represents the arguments needed to create a EnvironmentStatus.
type NewEnvironmentStatusParams struct {
	StatusID           string
	State              string
	LastDeploymentDate *time.Time
}

// NewStatus returns a new instance of a Status domain model.
func NewStatus(params NewStatusParams) (*Status, error) {
	statusUUID, err := uuid.Parse(params.StatusID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidStatusIDParam.Error())
	}

	state, err := NewStateFromString(params.State)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidStateParam.Error())
	}

	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidServiceDeploymentStatusParam.Error())
	}

	if params.LastDeploymentDate != nil && params.LastDeploymentDate.IsZero() {
		return nil, ErrInvalidLastDeploymentDateParam
	}

	s := &Status{
		ID:                 statusUUID,
		State:              *state,
		LastDeploymentDate: params.LastDeploymentDate,
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
}
