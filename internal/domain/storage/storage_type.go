package storage

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Type is an enum that contains all the valid values of a storage type.
type Type string

const (
	TypeFastSSD Type = "FAST_SSD"
)

// AllowedTypeValues contains all the valid values of a Type.
var AllowedTypeValues = []Type{
	TypeFastSSD,
}

// String returns the string value of a Type.
func (v Type) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Type is valid or not.
func (v Type) Validate() error {
	if slices.Contains(AllowedTypeValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Type: valid values are %v", v, AllowedTypeValues)
}

// IsValid returns a bool to tell whether the Type is valid or not.
func (v Type) IsValid() bool {
	return v.Validate() == nil
}

// NewTypeFromString tries to turn a string into a Type.
// It returns an error if the string is not a valid value.
func NewTypeFromString(v string) (*Type, error) {
	ev := Type(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
