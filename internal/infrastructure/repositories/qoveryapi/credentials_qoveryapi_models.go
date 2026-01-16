package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// newDomainCredentialsFromQovery takes a qovery.ClusterCredentials returned by the API client and turns it into the domain model credentials.Credentials.
func newDomainCredentialsFromQovery(organizationID string, creds *qovery.ClusterCredentials) (*credentials.Credentials, error) {
	if creds == nil {
		return nil, credentials.ErrNilCredentials
	}
	switch castedCreds := creds.GetActualInstance().(type) {
	case *qovery.AwsStaticClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	case *qovery.AwsRoleClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	case *qovery.ScalewayClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	case *qovery.GcpStaticClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	case *qovery.AzureStaticClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	case *qovery.GenericClusterCredentials:
		return credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
	default:
		return nil, errors.New("unknown credentials type")
	}
}

// newQoveryAwsCredentialsRequestFromDomain takes the domain request credentials.UpsertAwsRequest and turns it into a qovery.AwsCredentialsRequest to make the api call.
func newQoveryAwsCredentialsRequestFromDomain(request credentials.UpsertAwsRequest) qovery.AwsCredentialsRequest {
	req := qovery.AwsCredentialsRequest{}

	if creds := request.StaticCredentials; creds != nil {
		req.AwsStaticCredentialsRequest = &qovery.AwsStaticCredentialsRequest{
			Name:            request.Name,
			AccessKeyId:     creds.AccessKeyID,
			SecretAccessKey: creds.SecretAccessKey,
		}
	}
	if creds := request.RoleCredentials; creds != nil {
		req.AwsRoleCredentialsRequest = &qovery.AwsRoleCredentialsRequest{
			Name:    request.Name,
			RoleArn: creds.RoleArn,
		}
	}

	return req
}

// newQoveryScalewayCredentialsRequestFromDomain takes the domain request credentials.UpsertScalewayRequest and turns it into a qovery.ScalewayCredentialsRequest to make the api call.
func newQoveryScalewayCredentialsRequestFromDomain(request credentials.UpsertScalewayRequest) qovery.ScalewayCredentialsRequest {
	return qovery.ScalewayCredentialsRequest{
		Name:                   request.Name,
		ScalewayProjectId:      request.ScalewayProjectID,
		ScalewayAccessKey:      request.ScalewayAccessKey,
		ScalewaySecretKey:      request.ScalewaySecretKey,
		ScalewayOrganizationId: request.ScalewayOrganizationID,
	}
}

// newQoveryGcpCredentialsRequestFromDomain takes the domain request credentials.UpsertGcpRequest and turns it into a qovery.GcpCredentialsRequest to make the api call.
func newQoveryGcpCredentialsRequestFromDomain(request credentials.UpsertGcpRequest) qovery.GcpCredentialsRequest {
	return qovery.GcpCredentialsRequest{
		Name:           request.Name,
		GcpCredentials: request.GcpCredentials,
	}
}

// newQoveryAzureCredentialsRequestFromDomain takes the domain request credentials.UpsertAzureRequest and turns it into a qovery.AzureCredentialsRequest to make the api call.
func newQoveryAzureCredentialsRequestFromDomain(request credentials.UpsertAzureRequest) qovery.AzureCredentialsRequest {
	return qovery.AzureCredentialsRequest{
		Name:                request.Name,
		AzureSubscriptionId: request.AzureSubscriptionId,
		AzureTenantId:       request.AzureTenantId,
	}
}

// newDomainAzureCredentialsFromQovery takes a qovery.ClusterCredentials returned by the API client and turns it into the domain model credentials.AzureCredentials.
func newDomainAzureCredentialsFromQovery(organizationID string, creds *qovery.ClusterCredentials) (*credentials.AzureCredentials, error) {
	if creds == nil {
		return nil, credentials.ErrNilCredentials
	}

	switch castedCreds := creds.GetActualInstance().(type) {
	case *qovery.AzureStaticClusterCredentials:
		baseCreds, err := credentials.NewCredentials(credentials.NewCredentialsParams{
			CredentialsID:  castedCreds.GetId(),
			OrganizationID: organizationID,
			Name:           castedCreds.GetName(),
		})
		if err != nil {
			return nil, err
		}

		return &credentials.AzureCredentials{
			Credentials:              *baseCreds,
			AzureSubscriptionId:      castedCreds.GetAzureSubscriptionId(),
			AzureTenantId:            castedCreds.GetAzureTenantId(),
			AzureApplicationId:       castedCreds.GetAzureApplicationId(),
			AzureApplicationObjectId: castedCreds.GetAzureApplicationObjectId(),
		}, nil
	default:
		return nil, errors.New("unexpected credentials type for azure")
	}
}
