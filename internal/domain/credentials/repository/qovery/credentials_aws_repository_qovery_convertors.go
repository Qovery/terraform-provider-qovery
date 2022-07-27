package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// convertQoveryCredentialsToDomain takes a qovery.ClusterCredentials returned by the API client and turns it into the domain model credentials.Credentials.
func convertQoveryCredentialsToDomain(organizationID string, creds *qovery.ClusterCredentials) (*credentials.Credentials, error) {
	if creds == nil {
		return nil, credentials.ErrNilCredentials
	}

	return credentials.NewCredentials(creds.GetId(), organizationID, creds.GetName())
}

// convertDomainUpsertAwsRequestToQovery takes the domain request credentials.UpsertAwsRequest and turns it into a qovery.AwsCredentialsRequest to make the api call.
func convertDomainUpsertAwsRequestToQovery(request credentials.UpsertAwsRequest) qovery.AwsCredentialsRequest {
	return qovery.AwsCredentialsRequest{
		Name:            request.Name,
		AccessKeyId:     &request.AccessKeyID,
		SecretAccessKey: &request.SecretAccessKey,
	}
}
