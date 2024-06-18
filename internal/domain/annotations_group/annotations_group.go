package annotations_group

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

var (
	ErrInvalidAnnotationsGroupRequest             = errors.New("invalid annotations group request")
	ErrInvalidAnnotationsGroupOrganizationIdParam = errors.New("invalid organization id format")
	ErrFailedToCreateAnnotationsGroup             = errors.New("failed to create annotations group")
	ErrFailedToGetAnnotationsGroup                = errors.New("failed to get annotations group")
	ErrFailedToUpdateAnnotationsGroup             = errors.New("failed to update annotations group")
	ErrFailedToDeleteAnnotationsGroup             = errors.New("failed to delete annotations group")
	ErrInvalidAnnotationsGroupIdParam             = errors.New("invalid annotations group id format")
	ErrInvalidScope                               = errors.New("invalid scope")
)

type AnnotationsGroup struct {
	Id          uuid.UUID `validate:"required"`
	Name        string
	Annotations []qovery.Annotation
	Scopes      []qovery.OrganizationAnnotationsGroupScopeEnum
}
