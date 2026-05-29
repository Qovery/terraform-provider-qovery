package services

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// validateUUIDParam checks that value is a non-empty, well-formed UUID. It returns the supplied
// sentinel error (wrapped with parse context when the value is malformed) so callers keep their
// domain-specific error semantics while sharing the validation logic.
func validateUUIDParam(value string, sentinel error) error {
	if value == "" {
		return sentinel
	}
	if _, err := uuid.Parse(value); err != nil {
		return errors.Wrap(err, sentinel.Error())
	}
	return nil
}
