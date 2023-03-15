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
	MoveAfter     types.String `tfsdk:"move_after"`
	MoveBefore    types.String `tfsdk:"move_before"`
}

func (p DeploymentStage) toCreateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        toString(p.Name),
			Description: toString(p.Description),
			MoveAfter:   toStringPointer(p.MoveAfter),
			MoveBefore:  toStringPointer(p.MoveBefore),
		},
	}
}

func (p DeploymentStage) toUpdateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        toString(p.Name),
			Description: toString(p.Description),
			MoveAfter:   toStringPointer(p.MoveAfter),
			MoveBefore:  toStringPointer(p.MoveBefore),
		},
	}
}

func convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain *deploymentstage.DeploymentStage) DeploymentStage {
	var moveAfterString *string = nil
	if deploymentStageDomain.MoveAfter != nil {
		s := deploymentStageDomain.MoveAfter.String()
		moveAfterString = &s
	}
	var moveBeforeString *string = nil
	if deploymentStageDomain.MoveBefore != nil {
		s := deploymentStageDomain.MoveBefore.String()
		moveBeforeString = &s
	}
	return DeploymentStage{
		Id:            fromString(deploymentStageDomain.ID.String()),
		EnvironmentId: fromString(deploymentStageDomain.EnvironmentID.String()),
		Name:          fromString(deploymentStageDomain.Name),
		Description:   fromString(deploymentStageDomain.Description),
		MoveAfter:     fromStringPointer(moveAfterString),
		MoveBefore:    fromStringPointer(moveBeforeString),
	}
}
