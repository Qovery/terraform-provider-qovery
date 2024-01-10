package helmRepository

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateHelmRepository = errors.New("failed to create helm repository")
	ErrFailedToGetHelmRepository    = errors.New("failed to get helm repository")
	ErrFailedToUpdateHelmRepository = errors.New("failed to update helm repository")
	ErrFailedToDeleteHelmRepository = errors.New("failed to delete helm repository")
	ErrInvalidOrganizationIdParam   = errors.New("invalid organization Id")
	ErrInvalidRepositoryIdParam     = errors.New("invalid repository Id")
)

type Service interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*HelmRepository, error)
	Get(ctx context.Context, organizationID string, registryID string) (*HelmRepository, error)
	Update(ctx context.Context, organizationID string, registryID string, request UpsertRequest) (*HelmRepository, error)
	Delete(ctx context.Context, organizationID string, registryID string) error
}
