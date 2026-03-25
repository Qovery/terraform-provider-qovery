package common

import (
	"github.com/pkg/errors"
)

// ErrInvalidQoveryClient is the error return if the *qovery.Client is nil or invalid.
var ErrInvalidQoveryClient = errors.New("invalid qovery client")
