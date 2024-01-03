package helm

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Protocol string

const (
	ProtocolHttp Protocol = "HTTP"
	ProtocolGrpc Protocol = "GRPC"
)

var AllowedProtocols = []Protocol{
	ProtocolHttp,
	ProtocolGrpc,
}

func (v Protocol) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Kind is valid or not.
func (v Protocol) Validate() error {
	if slices.Contains(AllowedProtocols, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Protocol: valid values are %v", v, AllowedProtocols)
}

// IsValid returns a bool to tell whether the Kind is valid or not.
func (v Protocol) IsValid() bool {
	return v.Validate() == nil
}

func NewProtocolFromString(v string) (*Protocol, error) {
	ev := Protocol(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
