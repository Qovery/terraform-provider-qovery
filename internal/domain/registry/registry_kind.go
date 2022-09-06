package registry

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Kind is an enum that contains all the valid values of a registry kind.
type Kind string

const (
	KindECR        Kind = "ECR"
	KindDocker     Kind = "DOCR"
	KindScalewayCR Kind = "SCALEWAY_CR"
	KindDockerHub  Kind = "DOCKER_HUB"
	KindPublicECR  Kind = "PUBLIC_ECR"
)

// AllowedKindValues contains all the valid values of a Kind.
var AllowedKindValues = []Kind{
	KindECR,
	KindDocker,
	KindScalewayCR,
	KindDockerHub,
	KindPublicECR,
}

// String returns the string value of a Kind.
func (v Kind) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Kind is valid or not.
func (v Kind) Validate() error {
	if slices.Contains(AllowedKindValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Kind: valid values are %v", v, AllowedKindValues)
}

// IsValid returns a bool to tell whether the Kind is valid or not.
func (v Kind) IsValid() bool {
	return v.Validate() == nil
}

// NewKindFromString tries to turn a string into a Kind.
// It returns an error if the string is not a valid value.
func NewKindFromString(v string) (*Kind, error) {
	ev := Kind(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
