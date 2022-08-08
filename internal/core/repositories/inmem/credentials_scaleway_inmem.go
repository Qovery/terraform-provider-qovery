package inmem

import (
	"context"

	"github.com/google/uuid"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsScalewayInmem implements the interface credentials.ScalewayRepository.
type credentialsScalewayInmem struct {
	credentials map[string]*credentials.Credentials
}

// NOTE: This forces the implementation of the interface credentials.ScalewayRepository  by credentialsScalewayQoveryAPI at compile time.
var _ credentials.ScalewayRepository = credentialsScalewayInmem{}

// NewCredentialsScalewayInmem return a new instance of a credentials.ScalewayRepository that uses local memory storage.
func NewCredentialsScalewayInmem() credentials.ScalewayRepository {
	return credentialsScalewayInmem{
		credentials: make(map[string]*credentials.Credentials),
	}
}

// Create store in memory an aws cluster credentials on an organization using the given organizationID and request.
func (a credentialsScalewayInmem) Create(_ context.Context, organizationID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
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
func (a credentialsScalewayInmem) Get(_ context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, ok := a.credentials[credentialsID]
	if !ok || creds.OrganizationID.String() != organizationID {
		return nil, credentials.ErrScalewayCredentialsNotFound
	}

	return creds, nil
}

// Update updates in memory an aws cluster credentials from an organization using the given organizationID, credentialsID and request.
func (a credentialsScalewayInmem) Update(_ context.Context, organizationID string, credentialsID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	creds, ok := a.credentials[credentialsID]
	if !ok || creds.OrganizationID.String() != organizationID {
		return nil, credentials.ErrScalewayCredentialsNotFound
	}

	creds.Name = request.Name

	return creds, nil
}

// Delete deletes from memory an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (a credentialsScalewayInmem) Delete(_ context.Context, organizationID string, credentialsID string) error {
	delete(a.credentials, credentialsID)

	return nil
}
