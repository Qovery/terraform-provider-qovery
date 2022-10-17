package port

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Protocol is an enum that contains all the valid values of a port protocol.
type Protocol string

const (
	ProtocolHTTP Protocol = "HTTP"
)

// AllowedProtocolValues contains all the valid values of a Protocol.
var AllowedProtocolValues = []Protocol{
	ProtocolHTTP,
}

// String returns the string value of a Protocol.
func (v Protocol) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Protocol is valid or not.
func (v Protocol) Validate() error {
	if slices.Contains(AllowedProtocolValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Protocol: valid values are %v", v, AllowedProtocolValues)
}

// IsValid returns a bool to tell whether the Protocol is valid or not.
func (v Protocol) IsValid() bool {
	return v.Validate() == nil
}

// NewProtocolFromString tries to turn a string into a Protocol.
// It returns an error if the string is not a valid value.
func NewProtocolFromString(v string) (*Protocol, error) {
	ev := Protocol(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
