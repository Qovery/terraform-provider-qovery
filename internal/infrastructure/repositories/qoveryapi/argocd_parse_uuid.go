package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// parseUUID parses value as a UUID, wrapping any failure with the supplied sentinel error so the
// caller keeps its domain-specific error semantics. It centralises the parse-and-wrap block that
// the ArgoCD response converters would otherwise repeat for every UUID field.
func parseUUID(value string, sentinel error) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, sentinel.Error())
	}
	return parsed, nil
}
