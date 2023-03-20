package common

import (
	"github.com/pkg/errors"
)

var (
	// ErrInvalidQoveryClient is the error return if the *qovery.Client is nil or invalid.
	ErrInvalidQoveryClient = errors.New("invalid qovery client")
)
