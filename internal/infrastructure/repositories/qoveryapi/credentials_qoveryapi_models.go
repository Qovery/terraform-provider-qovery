package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// newDomainCredentialsFromQovery takes a qovery.ClusterCredentials returned by the API client and turns it into the domain model credentials.Credentials.
func newDomainCredentialsFromQovery(organizationID string, creds *qovery.ClusterCredentials) (*credentials.Credentials, error) {
	if creds == nil {
		return nil, credentials.ErrNilCredentials
	}

	return credentials.NewCredentials(credentials.NewCredentialsParams{
		CredentialsID:  creds.GetId(),
		OrganizationID: organizationID,
		Name:           creds.GetName(),
	})
}

// newQoveryAwsCredentialsRequestFromDomain takes the domain request credentials.UpsertAwsRequest and turns it into a qovery.AwsCredentialsRequest to make the api call.
func newQoveryAwsCredentialsRequestFromDomain(request credentials.UpsertAwsRequest) qovery.AwsCredentialsRequest {
	return qovery.AwsCredentialsRequest{
		Name:            request.Name,
		AccessKeyId:     request.AccessKeyID,
		SecretAccessKey: request.SecretAccessKey,
	}
}

// newQoveryScalewayCredentialsRequestFromDomain takes the domain request credentials.UpsertScalewayRequest and turns it into a qovery.ScalewayCredentialsRequest to make the api call.
func newQoveryScalewayCredentialsRequestFromDomain(request credentials.UpsertScalewayRequest) qovery.ScalewayCredentialsRequest {
	return qovery.ScalewayCredentialsRequest{
		Name:              request.Name,
		ScalewayProjectId: &request.ScalewayProjectID,
		ScalewayAccessKey: &request.ScalewayAccessKey,
		ScalewaySecretKey: &request.ScalewaySecretKey,
	}
}
