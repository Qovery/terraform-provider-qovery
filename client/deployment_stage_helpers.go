package client

import (
	"context"
	"net/http"

	"github.com/qovery/qovery-client-go"
)

// getServiceIsSkipped returns the is_skipped value for a specific service in a deployment stage response.
// It searches the services list for the entry matching the given serviceID.
// Returns false if the service is not found or is_skipped is not set.
func getServiceIsSkipped(deploymentStage *qovery.DeploymentStageResponse, serviceID string) bool {
	if deploymentStage == nil {
		return false
	}
	for _, svc := range deploymentStage.GetServices() {
		if svc.GetServiceId() == serviceID {
			return svc.GetIsSkipped()
		}
	}
	return false
}

// attachServiceToDeploymentStage attaches a service to a deployment stage with the given isSkipped flag.
// Returns the HTTP response and any error from the API call.
func attachServiceToDeploymentStage(ctx context.Context, api *qovery.APIClient, deploymentStageID string, serviceID string, isSkipped bool) (*http.Response, error) {
	attachRequest := qovery.NewAttachServiceToDeploymentStageRequest()
	attachRequest.SetIsSkipped(isSkipped)
	_, response, err := api.DeploymentStageMainCallsAPI.
		AttachServiceToDeploymentStage(ctx, deploymentStageID, serviceID).
		AttachServiceToDeploymentStageRequest(*attachRequest).
		Execute()
	return response, err
}
