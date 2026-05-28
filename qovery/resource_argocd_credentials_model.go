package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

type ArgoCdCredentials struct {
	Id          types.String `tfsdk:"id"`
	ClusterId   types.String `tfsdk:"cluster_id"`
	ArgocdUrl   types.String `tfsdk:"argocd_url"`
	ArgocdToken types.String `tfsdk:"argocd_token"`
}

func (a ArgoCdCredentials) toUpsertRequest() argoCdCredentials.UpsertRequest {
	return argoCdCredentials.UpsertRequest{
		ArgocdUrl:   ToString(a.ArgocdUrl),
		ArgocdToken: ToString(a.ArgocdToken),
	}
}

func convertDomainArgoCdCredentialsToTF(state ArgoCdCredentials, res *argoCdCredentials.ArgoCdCredentials) ArgoCdCredentials {
	return ArgoCdCredentials{
		Id:          FromString(res.ID.String()),
		ClusterId:   FromString(res.ClusterID.String()),
		ArgocdUrl:   FromString(res.ArgocdUrl),
		ArgocdToken: state.ArgocdToken, // API always returns REDACTED — preserve from state
	}
}
