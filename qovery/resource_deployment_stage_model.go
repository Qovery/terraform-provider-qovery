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
	IsAfter       types.String `tfsdk:"is_after"`
	IsBefore      types.String `tfsdk:"is_before"`
}

func (p DeploymentStage) toCreateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToString(p.Description),
			IsAfter:     ToStringPointer(p.IsAfter),
			IsBefore:    ToStringPointer(p.IsBefore),
		},
	}
}

func (p DeploymentStage) toUpdateServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        ToString(p.Name),
			Description: ToString(p.Description),
			IsAfter:     ToStringPointer(p.IsAfter),
			IsBefore:    ToStringPointer(p.IsBefore),
		},
	}
}

func convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain *deploymentstage.DeploymentStage, terraformDescription types.String) DeploymentStage {
	var isAfterString *string = nil
	if deploymentStageDomain.IsAfter != nil {
		s := deploymentStageDomain.IsAfter.String()
		isAfterString = &s
	}
	var isBeforeString *string = nil
	if deploymentStageDomain.IsBefore != nil {
		s := deploymentStageDomain.IsBefore.String()
		isBeforeString = &s
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
		IsAfter:       FromStringPointer(isAfterString),
		IsBefore:      FromStringPointer(isBeforeString),
	}
}
