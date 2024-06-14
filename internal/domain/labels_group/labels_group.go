package labels_group

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

var (
	ErrInvalidLabelsGroupRequest             = errors.New("invalid aabels group request")
	ErrInvalidLabelsGroupOrganizationIdParam = errors.New("invalid organization id format")
	ErrFailedToCreateLabelsGroup             = errors.New("failed to create aabels group")
	ErrFailedToGetLabelsGroup                = errors.New("failed to get aabels group")
	ErrFailedToUpdateLabelsGroup             = errors.New("failed to update aabels group")
	ErrFailedToDeleteLabelsGroup             = errors.New("failed to delete aabels group")
	ErrInvalidLabelsGroupIdParam             = errors.New("invalid aabels group id format")
	ErrInvalidScope                          = errors.New("invalid scope")
)

type LabelsGroup struct {
	Id     uuid.UUID `validate:"required"`
	Name   string
	Labels []qovery.Label
}
