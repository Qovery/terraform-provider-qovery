package environment

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Mode is an enum that contains all the valid values of an environment plan.
type Mode string

const (
	ModeDevelopment Mode = "DEVELOPMENT"
	ModePreview     Mode = "PREVIEW"
	ModeProduction  Mode = "PRODUCTION"
	ModeStaging     Mode = "STAGING"
)

// AllowedModeValues contains all the valid values of a Mode.
var AllowedModeValues = []Mode{
	ModeDevelopment,
	ModePreview,
	ModeProduction,
	ModeStaging,
}

// String returns the string value of a Mode.
func (v Mode) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Mode is valid or not.
func (v Mode) Validate() error {
	if slices.Contains(AllowedModeValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Mode: valid values are %v", v, AllowedModeValues)
}

// IsValid returns a bool to tell whether the Mode is valid or not.
func (v Mode) IsValid() bool {
	return v.Validate() == nil
}

// NewModeFromString tries to turn a string into a Mode.
// It returns an error if the string is not a valid value.
func NewModeFromString(v string) (*Mode, error) {
	ev := Mode(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
