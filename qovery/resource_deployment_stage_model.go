package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

type DeploymentStage struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
}

func (p DeploymentStage) toCreateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        toString(p.Name),
			Description: toString(p.Description),
		},
	}
}

func (p DeploymentStage) toUpdateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        toString(p.Name),
			Description: toString(p.Description),
		},
	}
}

func convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain *deploymentstage.DeploymentStage) DeploymentStage {
	return DeploymentStage{
		Id:            fromString(deploymentStageDomain.ID.String()),
		EnvironmentId: fromString(deploymentStageDomain.EnvironmentID.String()),
		Name:          fromString(deploymentStageDomain.Name),
		Description:   fromString(deploymentStageDomain.Description),
	}
}
