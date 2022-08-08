package services

import (
	"github.com/pkg/errors"
)

var (
	// ErrInvalidRepository is the error return if the given repository is nil or invalid.
	ErrInvalidRepository = errors.New("invalid repository")
)
