package deploymentrestriction

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type ServiceDeploymentRestrictionsDiff struct {
	Create []ServiceDeploymentRestriction
	Update []ServiceDeploymentRestriction
	Delete []string
}

func (d ServiceDeploymentRestrictionsDiff) IsNotEmpty() bool {
	return len(d.Create) > 0 || len(d.Update) > 0 || len(d.Delete) > 0
}

type ServiceDeploymentRestriction struct {
	Id    *string
	Mode  qovery.DeploymentRestrictionModeEnum
	Type  qovery.DeploymentRestrictionTypeEnum
	Value string
}

func ToDeploymentRestrictionDiff(deploymentRestrictionsSet types.Set, deploymentRestrictionsState *types.Set) (*ServiceDeploymentRestrictionsDiff, error) {
	if deploymentRestrictionsSet.IsNull() || deploymentRestrictionsSet.IsUnknown() {
		return &ServiceDeploymentRestrictionsDiff{
			Create: make([]ServiceDeploymentRestriction, 0),
			Update: make([]ServiceDeploymentRestriction, 0),
			Delete: make([]string, 0),
		}, nil
	}

	var deploymentRestrictionsToDelete = map[string]bool{}
	if deploymentRestrictionsState != nil && !deploymentRestrictionsState.IsNull() {
		for _, elem := range deploymentRestrictionsState.Elements() {
			elemToObject := elem.(types.Object)
			idStr := elemToObject.Attributes()["id"].(types.String).ValueString()
			deploymentRestrictionsToDelete[idStr] = true
		}
	}

	// deployment restriction with no id will be created
	toCreate := make([]ServiceDeploymentRestriction, 0)
	// deployment restriction with id will be updated
	toUpdate := make([]ServiceDeploymentRestriction, 0)

	for _, elem := range deploymentRestrictionsSet.Elements() {
		elemToObject := elem.(types.Object)
		id := elemToObject.Attributes()["id"].(types.String)
		modeStr := elemToObject.Attributes()["mode"].(types.String).ValueString()
		typeStr := elemToObject.Attributes()["type"].(types.String).ValueString()
		valueStr := elemToObject.Attributes()["value"].(types.String).ValueString()

		modeEnum, err := qovery.NewDeploymentRestrictionModeEnumFromValue(modeStr)
		if err != nil {
			return nil, err
		}

		typeEnum, err := qovery.NewDeploymentRestrictionTypeEnumFromValue(typeStr)
		if err != nil {
			return nil, err
		}

		if id.IsNull() || id.IsUnknown() {
			toCreate = append(toCreate, ServiceDeploymentRestriction{
				Id:    nil,
				Mode:  *modeEnum,
				Type:  *typeEnum,
				Value: valueStr,
			})
		} else {
			idStr := id.ValueString()
			toUpdate = append(toUpdate, ServiceDeploymentRestriction{
				Id:    &idStr,
				Mode:  *modeEnum,
				Type:  *typeEnum,
				Value: valueStr,
			})

			delete(deploymentRestrictionsToDelete, idStr)
		}
	}

	toDelete := make([]string, len(deploymentRestrictionsToDelete))

	i := 0
	for deploymentRestrictionIdToDelete := range deploymentRestrictionsToDelete {
		toDelete[i] = deploymentRestrictionIdToDelete
		i++
	}

	return &ServiceDeploymentRestrictionsDiff{
		Create: toCreate,
		Update: toUpdate,
		Delete: toDelete,
	}, nil
}
