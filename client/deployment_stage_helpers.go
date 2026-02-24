package client

import (
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
