package inmem

import (
	"context"

	"github.com/google/uuid"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAwsInmem implements the interface credentials.AwsRepository.
type credentialsAwsInmem struct {
	credentials map[string]*credentials.Credentials
}

// NOTE: This forces the implementation of the interface credentials.AwsRepository  by credentialsAwsQoveryAPI at compile time.
var _ credentials.AwsRepository = credentialsAwsInmem{}

// NewCredentialsAwsInmem return a new instance of a credentials.AwsRepository that uses local memory storage.
func NewCredentialsAwsInmem() credentials.AwsRepository {
	return credentialsAwsInmem{
		credentials: make(map[string]*credentials.Credentials),
	}
}

// Create store in memory an aws cluster credentials on an organization using the given organizationID and request.
func (a credentialsAwsInmem) Create(_ context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, err := credentials.NewCredentials(credentials.NewCredentialsParams{
		CredentialsID:  uuid.NewString(),
		OrganizationID: organizationID,
		Name:           request.Name,
	})
	if err != nil {
		return nil, err
	}

	a.credentials[creds.ID.String()] = creds

	return creds, nil
}

// Get retrieve from memory an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (a credentialsAwsInmem) Get(_ context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, ok := a.credentials[credentialsID]
	if !ok || creds.OrganizationID.String() != organizationID {
		return nil, credentials.ErrAwsCredentialsNotFound
	}

	return creds, nil
}

// Update updates in memory an aws cluster credentials from an organization using the given organizationID, credentialsID and request.
func (a credentialsAwsInmem) Update(_ context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, ok := a.credentials[credentialsID]
	if !ok || creds.OrganizationID.String() != organizationID {
		return nil, credentials.ErrAwsCredentialsNotFound
	}

	creds.Name = request.Name

	return creds, nil
}

// Delete deletes from memory an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (a credentialsAwsInmem) Delete(_ context.Context, organizationID string, credentialsID string) error {
	delete(a.credentials, credentialsID)

	return nil
}
