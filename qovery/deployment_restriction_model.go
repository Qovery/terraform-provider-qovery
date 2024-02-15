package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
)

var deploymentRestrictionsAttrTypes = map[string]attr.Type{
	"id":    types.StringType,
	"mode":  types.StringType,
	"type":  types.StringType,
	"value": types.StringType,
}

type DeploymentRestriction struct {
	Id    types.String `tfsdk:"id"`
	Mode  types.String `tfsdk:"mode"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

var deploymentRestrictionObjectType = types.ObjectType{
	AttrTypes: deploymentRestrictionsAttrTypes,
}

type DeploymentRestrictionList []DeploymentRestriction

func FromDeploymentRestrictionList(initialState types.Set, deploymentRestrictions []deploymentrestriction.ServiceDeploymentRestriction) types.Set {
	var list DeploymentRestrictionList
	list = make([]DeploymentRestriction, 0, len(deploymentRestrictions))
	for _, deploymentRestriction := range deploymentRestrictions {
		modeStr := fmt.Sprintf("%s", deploymentRestriction.Mode)
		typeStr := fmt.Sprintf("%s", deploymentRestriction.Type)
		list = append(list, DeploymentRestriction{
			Id:    FromString(*deploymentRestriction.Id),
			Mode:  FromString(modeStr),
			Type:  FromString(typeStr),
			Value: FromString(deploymentRestriction.Value),
		})
	}

	if len(list) == 0 && initialState.IsNull() {
		return types.SetNull(deploymentRestrictionObjectType)
	}

	if len(initialState.Elements()) == 0 {
		return types.SetValueMust(deploymentRestrictionObjectType, []attr.Value{})
	}

	var elements = make([]attr.Value, 0, len(list))
	for _, v := range list {
		elements = append(elements, v.toTerraformObject())
	}

	return types.SetValueMust(deploymentRestrictionObjectType, elements)
}

func (dr DeploymentRestriction) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"id":    dr.Id,
		"mode":  dr.Mode,
		"type":  dr.Type,
		"value": dr.Value,
	}
	terraformObjectValue := types.ObjectValueMust(deploymentRestrictionsAttrTypes, attributes)

	return terraformObjectValue
}
