package qoveryapi

import (
	"context"
	"fmt"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

var _ argoCdCredentials.Repository = argoCdCredentialsQoveryAPI{}

type argoCdCredentialsQoveryAPI struct {
	client *qovery.APIClient
}

func newArgoCdCredentialsQoveryAPI(client *qovery.APIClient) (argoCdCredentials.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &argoCdCredentialsQoveryAPI{client: client}, nil
}

func (a argoCdCredentialsQoveryAPI) Create(ctx context.Context, clusterID string, request argoCdCredentials.UpsertRequest) (*argoCdCredentials.ArgoCdCredentials, error) {
	req := qovery.NewArgoCdCredentialsRequest(request.ArgocdUrl, request.ArgocdToken)
	if err := a.checkConnection(ctx, clusterID, req); err != nil {
		return nil, err
	}
	res, resp, err := a.client.ArgoCDAPI.
		SaveArgoCdCredentials(ctx, clusterID).
		ArgoCdCredentialsRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceArgoCdCredentials, clusterID, resp, err)
	}
	return newDomainArgoCdCredentialsFromQovery(res)
}

func (a argoCdCredentialsQoveryAPI) Get(ctx context.Context, clusterID string) (*argoCdCredentials.ArgoCdCredentials, error) {
	res, resp, err := a.client.ArgoCDAPI.
		GetArgoCdCredentials(ctx, clusterID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceArgoCdCredentials, clusterID, resp, err)
	}
	return newDomainArgoCdCredentialsFromQovery(res)
}

func (a argoCdCredentialsQoveryAPI) Update(ctx context.Context, clusterID string, request argoCdCredentials.UpsertRequest) (*argoCdCredentials.ArgoCdCredentials, error) {
	req := qovery.NewArgoCdCredentialsRequest(request.ArgocdUrl, request.ArgocdToken)
	if err := a.checkConnection(ctx, clusterID, req); err != nil {
		return nil, err
	}
	res, resp, err := a.client.ArgoCDAPI.
		SaveArgoCdCredentials(ctx, clusterID).
		ArgoCdCredentialsRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceArgoCdCredentials, clusterID, resp, err)
	}
	return newDomainArgoCdCredentialsFromQovery(res)
}

func (a argoCdCredentialsQoveryAPI) Delete(ctx context.Context, clusterID string) error {
	resp, err := a.client.ArgoCDAPI.
		DeleteArgoCdCredentials(ctx, clusterID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceArgoCdCredentials, clusterID, resp, err)
	}
	return nil
}

func (a argoCdCredentialsQoveryAPI) checkConnection(ctx context.Context, clusterID string, req *qovery.ArgoCdCredentialsRequest) error {
	checkRes, resp, err := a.client.ArgoCDAPI.
		CheckArgoCdConnection(ctx, clusterID).
		ArgoCdCredentialsRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return apierrors.NewCreateAPIError(apierrors.APIResourceArgoCdCredentials, clusterID, resp, err)
	}
	if checkRes.Status == qovery.ARGOCDCONNECTIONSTATUSENUM_ERROR {
		reason := ""
		if checkRes.Reason != nil {
			reason = *checkRes.Reason
		}
		return fmt.Errorf("argocd connection check failed: %s", reason)
	}
	return nil
}

func newDomainArgoCdCredentialsFromQovery(res *qovery.ArgoCdCredentialsResponse) (*argoCdCredentials.ArgoCdCredentials, error) {
	id, err := parseUUID(res.Id, argoCdCredentials.ErrInvalidClusterIDParam)
	if err != nil {
		return nil, err
	}
	clusterID, err := parseUUID(res.ClusterId, argoCdCredentials.ErrInvalidClusterIDParam)
	if err != nil {
		return nil, err
	}
	return &argoCdCredentials.ArgoCdCredentials{
		ID:          id,
		ClusterID:   clusterID,
		ArgocdUrl:   res.ArgocdUrl,
		ArgocdToken: res.ArgocdToken,
	}, nil
}
