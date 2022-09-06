package variable

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Scope is an enum that contains all the valid values of a variable scope.
type Scope string

const (
	ScopeApplication Scope = "APPLICATION"
	ScopeBuiltIn     Scope = "BUILT_IN"
	ScopeContainer   Scope = "CONTAINER"
	ScopeEnvironment Scope = "ENVIRONMENT"
	ScopeProject     Scope = "PROJECT"
)

// AllowedScopeValues contains all the valid values of a Scope.
var AllowedScopeValues = []Scope{
	ScopeApplication,
	ScopeBuiltIn,
	ScopeContainer,
	ScopeEnvironment,
	ScopeProject,
}

// String returns the string value of a Scope.
func (v Scope) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Scope is valid or not.
func (v Scope) Validate() error {
	if slices.Contains(AllowedScopeValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Scope: valid values are %v", v, AllowedScopeValues)
}

// IsValid returns a bool to tell whether the Scope is valid or not.
func (v Scope) IsValid() bool {
	return v.Validate() == nil
}

// NewScopeFromString tries to turn a string into a Scope.
// It returns an error if the string is not a valid value.
func NewScopeFromString(v string) (*Scope, error) {
	ev := Scope(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
