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
			Name:        ToString(p.Name),
			Description: ToString(p.Description),
			MoveAfter:   ToStringPointer(p.MoveAfter),
			MoveBefore:  ToStringPointer(p.MoveBefore),
		},
	}
}

func (p DeploymentStage) toUpdateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToString(p.Description),
			MoveAfter:   ToStringPointer(p.MoveAfter),
			MoveBefore:  ToStringPointer(p.MoveBefore),
		},
	}
}

func convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain *deploymentstage.DeploymentStage, terraformDescription types.String) DeploymentStage {
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

	// hack to satisfy optional description as the core doesn't accept null value, the description will be always an empty string in case of null
	var description = &deploymentStageDomain.Description
	if deploymentStageDomain.Description == "" && terraformDescription.IsNull() {
		description = nil
	}

	return DeploymentStage{
		Id:            FromString(deploymentStageDomain.ID.String()),
		EnvironmentId: FromString(deploymentStageDomain.EnvironmentID.String()),
		Name:          FromString(deploymentStageDomain.Name),
		Description:   FromStringPointer(description),
		MoveAfter:     FromStringPointer(moveAfterString),
		MoveBefore:    FromStringPointer(moveBeforeString),
	}
}
