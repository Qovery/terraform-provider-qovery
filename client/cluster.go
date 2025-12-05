package client

import (
	"context"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

type ClusterResponse struct {
	OrganizationID       string
	ClusterResponse      *qovery.Cluster
	ClusterInfo          *qovery.ClusterCloudProviderInfo
	ClusterRoutingTable  *ClusterRoutingTable
	AdvancedSettingsJson string
}

type ClusterUpsertParams struct {
	ClusterRequest              qovery.ClusterRequest
	ClusterCloudProviderRequest *qovery.ClusterCloudProviderInfoRequest
	ClusterRoutingTable         ClusterRoutingTable
	AdvancedSettingsJson        string
	ForceUpdate                 bool
	DesiredState                qovery.ClusterStateEnum
}

func (c *Client) CreateCluster(ctx context.Context, organizationID string, params *ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	cluster, res, err := c.api.ClustersAPI.
		CreateCluster(ctx, organizationID).
		ClusterRequest(params.ClusterRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceCluster, params.ClusterRequest.Name, res, err)
	}
	return c.updateCluster(ctx, organizationID, cluster, params)
}

func (c *Client) GetCluster(ctx context.Context, organizationID string, clusterID string, advancedSettingsFromState string, isTriggeredFromImport bool) (*ClusterResponse, *apierrors.APIError) {
	cluster, apiErr := c.getClusterByID(ctx, organizationID, clusterID)
	if apiErr != nil {
		return nil, apiErr
	}

	clusterInfo, res, err := c.api.ClustersAPI.
		GetOrganizationCloudProviderInfo(ctx, organizationID, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
	}

	clusterRoutingTable, apiErr := c.getClusterRoutingTable(ctx, organizationID, clusterID)
	if apiErr != nil {
		return nil, apiErr
	}

	advancedSettingsJson, err := advanced_settings.NewClusterAdvancedSettingsService(c.api.GetConfig()).ReadClusterAdvancedSettings(organizationID, clusterID, advancedSettingsFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterAdvancedSettings, cluster.Id, nil, err)
	}

	return &ClusterResponse{
		OrganizationID:       organizationID,
		ClusterResponse:      cluster,
		ClusterRoutingTable:  clusterRoutingTable,
		ClusterInfo:          clusterInfo,
		AdvancedSettingsJson: *advancedSettingsJson,
	}, nil
}

func (c *Client) UpdateCluster(ctx context.Context, organizationID string, clusterID string, params *ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	// INFO (cor-775) As DiskSize is defaulted when no value is present in the request, we need to set it to current value
	// This is due to the attribute `disk_size` that was not there before
	// INFO: Same logic applies for MetricsParameters and InfrastructureChartsParameters - preserve value set via console without exposing it in Terraform state
	if params.ClusterRequest.DiskSize == nil || params.ClusterRequest.MetricsParameters == nil || params.ClusterRequest.InfrastructureChartsParameters == nil {
		cluster, apiErr := c.getClusterByID(ctx, organizationID, clusterID)
		if apiErr != nil {
			return nil, apiErr
		}
		if params.ClusterRequest.DiskSize == nil {
			params.ClusterRequest.DiskSize = cluster.DiskSize
		}
		if params.ClusterRequest.MetricsParameters == nil {
			params.ClusterRequest.MetricsParameters = cluster.MetricsParameters
		}
		if params.ClusterRequest.InfrastructureChartsParameters == nil {
			params.ClusterRequest.InfrastructureChartsParameters = cluster.InfrastructureChartsParameters
		}
	}
	cluster, res, err := c.api.ClustersAPI.
		EditCluster(ctx, organizationID, clusterID).
		ClusterRequest(params.ClusterRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	return c.updateCluster(ctx, organizationID, cluster, params)
}

func (c *Client) DeleteCluster(ctx context.Context, organizationID string, clusterID string) *apierrors.APIError {
	finalStateChecker := newClusterFinalStateCheckerWaitFunc(c, organizationID, clusterID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.api.ClustersAPI.
		DeleteCluster(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	checker := newClusterStatusCheckerWaitFunc(c, organizationID, clusterID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) getClusterByID(ctx context.Context, organizationID string, clusterID string) (*qovery.Cluster, *apierrors.APIError) {
	clusters, res, err := c.api.ClustersAPI.
		ListOrganizationCluster(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	for _, cluster := range clusters.GetResults() {
		if cluster.Id == clusterID {
			return &cluster, nil
		}
	}

	// NOTE: Force status 404 since we didn't find the credential.
	// The status is used to generate the proper error return by the provider.
	res.StatusCode = 404
	return nil, apierrors.NewReadError(apierrors.APIResourceCluster, clusterID, res, err)
}

func (c *Client) updateCluster(ctx context.Context, organizationID string, cluster *qovery.Cluster, params *ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	if params.ClusterCloudProviderRequest != nil {
		_, res, err := c.api.ClustersAPI.
			SpecifyClusterCloudProviderInfo(ctx, organizationID, cluster.Id).
			ClusterCloudProviderInfoRequest(*params.ClusterCloudProviderRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
		}
	}

	clusterInfo, res, err := c.api.ClustersAPI.
		GetOrganizationCloudProviderInfo(ctx, organizationID, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
	}

	var clusterRoutingTable *ClusterRoutingTable
	if len(params.ClusterRoutingTable.Routes) > 0 {
		var apiErr *apierrors.APIError
		clusterRoutingTable, apiErr = c.editClusterRoutingTable(ctx, organizationID, cluster.Id, params.ClusterRoutingTable)
		if apiErr != nil {
			return nil, apiErr
		}
	}

	err = advanced_settings.NewClusterAdvancedSettingsService(c.api.GetConfig()).UpdateClusterAdvancedSettings(organizationID, cluster.Id, params.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterAdvancedSettings, cluster.Id, nil, err)
	}

	clusterStatus, apiErr := c.updateClusterStatus(ctx, organizationID, cluster, params.DesiredState, params.ForceUpdate)
	if apiErr != nil {
		return nil, apiErr
	}
	cluster.Status = clusterStatus

	// Get cluster because after the updateClusterStatus (that may deploy the cluster), the InfrastructureOutputs can change.
	updatedCluster, apiErr := c.getClusterByID(ctx, organizationID, cluster.Id)
	if apiErr == nil {
		cluster.InfrastructureOutputs = updatedCluster.InfrastructureOutputs
	}

	return &ClusterResponse{
		OrganizationID:       organizationID,
		ClusterResponse:      cluster,
		ClusterRoutingTable:  clusterRoutingTable,
		ClusterInfo:          clusterInfo,
		AdvancedSettingsJson: params.AdvancedSettingsJson,
	}, nil
}

func (c *Client) GetClusterKubeconfig(ctx context.Context, organizationID string, clusterID string) (string, *apierrors.APIError) {
	kubeconfig, res, err := c.api.ClustersAPI.
		GetClusterKubeconfig(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return "", apierrors.NewReadError(apierrors.APIResourceCluster, clusterID, res, err)
	}
	return kubeconfig, nil
}

func (c *Client) SetClusterKubeconfig(ctx context.Context, organizationID string, clusterID string, kubeconfig string) *apierrors.APIError {
	res, err := c.api.ClustersAPI.
		EditClusterKubeconfig(ctx, organizationID, clusterID).
		Body(kubeconfig).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceCluster, clusterID, res, err)
	}
	return nil
}
