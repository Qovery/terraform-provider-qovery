package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// newDomainCredentialsFromQovery takes a qovery.ClusterCredentials returned by the API client and turns it into the domain model credentials.Credentials,
// with the identifiers the API returns for its type.
func newDomainCredentialsFromQovery(organizationID string, creds *qovery.ClusterCredentials) (*credentials.Credentials, error) {
	if creds == nil {
		return nil, credentials.ErrNilCredentials
	}
	params := credentials.NewCredentialsParams{OrganizationID: organizationID}
	switch castedCreds := creds.GetActualInstance().(type) {
	case *qovery.AwsStaticClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
		params.Identifiers.AccessKeyID = new(castedCreds.GetAccessKeyId())
	case *qovery.AwsRoleClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
		params.Identifiers.RoleArn = new(castedCreds.GetRoleArn())
	case *qovery.ScalewayClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
		params.Identifiers.ScalewayAccessKey = new(castedCreds.GetScalewayAccessKey())
		params.Identifiers.ScalewayProjectID = new(castedCreds.GetScalewayProjectId())
		params.Identifiers.ScalewayOrganizationID = new(castedCreds.GetScalewayOrganizationId())
	case *qovery.GcpStaticClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
	case *qovery.GcpWorkloadIdentityFederationClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
		params.Identifiers.ServiceAccountEmail = new(castedCreds.GetServiceAccountEmail())
		params.Identifiers.WorkloadIdentityProviderResource = new(castedCreds.GetWorkloadIdentityProviderResource())
	case *qovery.AzureStaticClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
	case *qovery.GenericClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
	case *qovery.EksAnywhereVsphereClusterCredentials:
		params.CredentialsID, params.Name = castedCreds.GetId(), castedCreds.GetName()
		params.Identifiers.VsphereUser = new(castedCreds.GetVsphereUser())
		params.Identifiers.AccessKeyID = castedCreds.AccessKeyId.Get()
		params.Identifiers.RoleArn = castedCreds.RoleArn.Get()
	default:
		return nil, errors.New("unknown credentials type")
	}
	return credentials.NewCredentials(params)
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
	if request.WorkloadIdentity != nil {
		return qovery.GcpWorkloadIdentityFederationCredentialsRequestAsGcpCredentialsRequest(&qovery.GcpWorkloadIdentityFederationCredentialsRequest{
			Name:                             request.Name,
			ServiceAccountEmail:              request.WorkloadIdentity.ServiceAccountEmail,
			WorkloadIdentityProviderResource: request.WorkloadIdentity.WorkloadIdentityProviderResource,
		})
	}

	return qovery.GcpServiceAccountKeyCredentialsRequestAsGcpCredentialsRequest(&qovery.GcpServiceAccountKeyCredentialsRequest{
		Name:           request.Name,
		GcpCredentials: request.ServiceAccountKey.GcpCredentials,
	})
}

// newQoveryAzureCredentialsRequestFromDomain takes the domain request credentials.UpsertAzureRequest and turns it into a qovery.AzureCredentialsRequest to make the api call.
func newQoveryAzureCredentialsRequestFromDomain(request credentials.UpsertAzureRequest) qovery.AzureCredentialsRequest {
	return qovery.AzureCredentialsRequest{
		Name:                request.Name,
		AzureSubscriptionId: request.AzureSubscriptionId,
		AzureTenantId:       request.AzureTenantId,
	}
}

// newQoveryEksAnywhereVsphereCredentialsRequestFromDomain takes the domain request credentials.UpsertEksAnywhereVsphereRequest and turns it into a qovery.AwsCredentialsRequest to make the api call.
func newQoveryEksAnywhereVsphereCredentialsRequestFromDomain(request credentials.UpsertEksAnywhereVsphereRequest) qovery.AwsCredentialsRequest {
	if creds := request.StaticCredentials; creds != nil {
		return qovery.EksAnywhereVsphereStaticCredentialsRequestAsAwsCredentialsRequest(&qovery.EksAnywhereVsphereStaticCredentialsRequest{
			Type:            "EKS_ANYWHERE_VSPHERE_STATIC",
			Name:            request.Name,
			VsphereUser:     request.VsphereUser,
			VspherePassword: request.VspherePassword,
			AccessKeyId:     creds.AccessKeyID,
			SecretAccessKey: creds.SecretAccessKey,
		})
	}
	return qovery.EksAnywhereVsphereRoleCredentialsRequestAsAwsCredentialsRequest(&qovery.EksAnywhereVsphereRoleCredentialsRequest{
		Type:            "EKS_ANYWHERE_VSPHERE_ROLE",
		Name:            request.Name,
		VsphereUser:     request.VsphereUser,
		VspherePassword: request.VspherePassword,
		RoleArn:         request.RoleCredentials.RoleArn,
	})
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
