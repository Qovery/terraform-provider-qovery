package helmRepository

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Kind is an enum that contains all the valid values of a registry kind.
type Kind string

const (
	KindHttps      Kind = "HTTPS"
	KindECR        Kind = "OCI_ECR"
	KindDocker     Kind = "OCI_DOCR"
	KindScalewayCR Kind = "OCI_SCALEWAY_CR"
	KindDockerHub  Kind = "OCI_DOCKER_HUB"
	KindGithubCr   Kind = "OCI_GITHUB_CR"
	KindGitlabCr   Kind = "OCI_GITLAB_CR"
	KindPublicECR  Kind = "OCI_PUBLIC_ECR"
	KindGenericCR  Kind = "OCI_GENERIC_CR"
)

// AllowedKindValues contains all the valid values of a Kind.
var AllowedKindValues = []Kind{
	KindHttps,
	KindECR,
	KindDocker,
	KindScalewayCR,
	KindDockerHub,
	KindGithubCr,
	KindGitlabCr,
	KindPublicECR,
	KindGenericCR,
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
