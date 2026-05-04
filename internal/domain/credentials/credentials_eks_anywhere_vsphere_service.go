package credentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateEksAnywhereVsphereCredentials = errors.New("failed to create eks anywhere vsphere credentials")
	ErrFailedToGetEksAnywhereVsphereCredentials    = errors.New("failed to get eks anywhere vsphere credentials")
	ErrFailedToUpdateEksAnywhereVsphereCredentials = errors.New("failed to update eks anywhere vsphere credentials")
	ErrFailedToDeleteEksAnywhereVsphereCredentials = errors.New("failed to delete eks anywhere vsphere credentials")
)

// EksAnywhereVsphereService represents the interface to implement to handle the domain logic of EKS Anywhere vSphere Credentials.
type EksAnywhereVsphereService interface {
	Create(ctx context.Context, organizationID string, request UpsertEksAnywhereVsphereRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertEksAnywhereVsphereRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
