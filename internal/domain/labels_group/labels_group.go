package labels_group

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

var (
	ErrInvalidLabelsGroupRequest             = errors.New("invalid labels group request")
	ErrInvalidLabelsGroupOrganizationIdParam = errors.New("invalid organization id format")
	ErrFailedToCreateLabelsGroup             = errors.New("failed to create labels group")
	ErrFailedToGetLabelsGroup                = errors.New("failed to get labels group")
	ErrFailedToUpdateLabelsGroup             = errors.New("failed to update labels group")
	ErrFailedToDeleteLabelsGroup             = errors.New("failed to delete labels group")
	ErrInvalidLabelsGroupIdParam             = errors.New("invalid labels group id format")
	ErrInvalidScope                          = errors.New("invalid scope")
)

type LabelsGroup struct {
	Id     uuid.UUID `validate:"required"`
	Name   string
	Labels []qovery.Label
}
