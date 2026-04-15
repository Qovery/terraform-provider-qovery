package credentials

//go:generate mockery --testonly --with-expecter --name=EksAnywhereVsphereRepository --structname=CredentialsEksAnywhereVsphereRepository --filename=credentials_eks_anywhere_vsphere_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// EksAnywhereVsphereRepository represents the interface to implement to handle the persistence of EKS Anywhere vSphere Credentials.
type EksAnywhereVsphereRepository interface {
	Create(ctx context.Context, organizationID string, request UpsertEksAnywhereVsphereRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertEksAnywhereVsphereRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
