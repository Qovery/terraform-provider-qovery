package status

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// State is an enum that contains all the valid values of a status protocol.
type State string

const (
	StateBuilding          State = "BUILDING"
	StateBuildError        State = "BUILD_ERROR"
	StateCanceled          State = "CANCELED"
	StateCanceling         State = "CANCELING"
	StateDeleted           State = "DELETED"
	StateDeleteError       State = "DELETE_ERROR"
	StateDeleteQueued      State = "DELETE_QUEUED"
	StateDeleting          State = "DELETING"
	StateDeployed          State = "DEPLOYED"
	StateDeploying         State = "DEPLOYING"
	StateDeploymentError   State = "DEPLOYMENT_ERROR"
	StateDeploymentQueued  State = "DEPLOYMENT_QUEUED"
	StateExecuting         State = "EXECUTING"
	StateQueued            State = "QUEUED"
	StateReady             State = "READY"
	StateStopped           State = "STOPPED"
	StateStopping          State = "STOPPING"
	StateStopError         State = "STOP_ERROR"
	StateStopQueued        State = "STOP_QUEUED"
	StateRestarting        State = "RESTARTING"
	StateRestarted         State = "RESTARTED"
	StateRestartError      State = "RESTART_ERROR"
	StateRestartQueued     State = "RESTART_QUEUED"
	StateRecap             State = "RECAP"
	StateWaitingDeleting   State = "WAITING_DELETING"
	StateWaitingRestarting State = "WAITING_RESTARTING"
	StateWaitingRunning    State = "WAITING_RUNNING"
	StateWaitingStopping   State = "WAITING_STOPPING"
)

// AllowedStateValues contains all the valid values of a State.
var AllowedStateValues = []State{
	StateBuilding,
	StateBuildError,
	StateCanceled,
	StateCanceling,
	StateDeleted,
	StateDeleteError,
	StateDeleteQueued,
	StateDeleting,
	StateDeployed,
	StateDeploying,
	StateDeploymentError,
	StateDeploymentQueued,
	StateExecuting,
	StateQueued,
	StateReady,
	StateStopped,
	StateStopping,
	StateStopError,
	StateStopQueued,
	StateRestarting,
	StateRestarted,
	StateRestartError,
	StateRestartQueued,
	StateRecap,
	StateWaitingDeleting,
	StateWaitingRestarting,
	StateWaitingRunning,
	StateWaitingStopping,
}

var AllowedDesiredStateValues = []State{
	StateDeployed,
	StateStopped,
	StateRestarted,
}

// String returns the string value of a State.
func (v State) String() string {
	return string(v)
}

// Validate returns an error to tell whether the State is valid or not.
func (v State) Validate() error {
	if slices.Contains(AllowedStateValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for State: valid values are %v", v, AllowedStateValues)
}

// IsValid returns a bool to tell whether the State is valid or not.
func (v State) IsValid() bool {
	return v.Validate() == nil
}

// NewStateFromString tries to turn a string into a State.
// It returns an error if the string is not a valid value.
func NewStateFromString(v string) (*State, error) {
	ev := State(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
