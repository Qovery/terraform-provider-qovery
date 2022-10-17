package status

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// ServiceDeploymentStatus is an enum that contains all the valid values of a status protocol.
type ServiceDeploymentStatus string

const (
	ServiceDeploymentStatusNeverDeployed ServiceDeploymentStatus = "NEVER_DEPLOYED"
	ServiceDeploymentStatusOutOfDate     ServiceDeploymentStatus = "OUT_OF_DATE"
	ServiceDeploymentStatusUpToDate      ServiceDeploymentStatus = "UP_TO_DATE"
)

// AllowedServiceDeploymentStatusValues contains all the valid values of a ServiceDeploymentStatus.
var AllowedServiceDeploymentStatusValues = []ServiceDeploymentStatus{
	ServiceDeploymentStatusNeverDeployed,
	ServiceDeploymentStatusOutOfDate,
	ServiceDeploymentStatusUpToDate,
}

// String returns the string value of a ServiceDeploymentStatus.
func (v ServiceDeploymentStatus) String() string {
	return string(v)
}

// Validate returns an error to tell whether the ServiceDeploymentStatus is valid or not.
func (v ServiceDeploymentStatus) Validate() error {
	if slices.Contains(AllowedServiceDeploymentStatusValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for ServiceDeploymentStatus: valid values are %v", v, AllowedServiceDeploymentStatusValues)
}

// IsValid returns a bool to tell whether the ServiceDeploymentStatus is valid or not.
func (v ServiceDeploymentStatus) IsValid() bool {
	return v.Validate() == nil
}

// NewServiceDeploymentStatusFromString tries to turn a string into a ServiceDeploymentStatus.
// It returns an error if the string is not a valid value.
func NewServiceDeploymentStatusFromString(v string) (*ServiceDeploymentStatus, error) {
	ev := ServiceDeploymentStatus(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
