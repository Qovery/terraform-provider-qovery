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

	elements := make([]attr.Value, 0, len(list))
	for _, v := range list {
		elements = append(elements, v.toTerraformObject())
	}

	// When the API returns restrictions, always reflect them (with their IDs) in
	// state. This is required for imported resources: their initial state has no
	// restrictions, so we must reconcile from the API by ID instead of trying to
	// recreate restrictions that already exist (which causes a 409 Conflict on the
	// next apply).
	if len(elements) == 0 {
		// No restrictions on the API side: preserve the prior null-vs-empty shape so
		// a never-configured block stays null while an explicitly empty set stays empty.
		if initialState.IsNull() {
			return types.SetNull(deploymentRestrictionObjectType)
		}
		return types.SetValueMust(deploymentRestrictionObjectType, []attr.Value{})
	}

	return types.SetValueMust(deploymentRestrictionObjectType, elements)
}

func (dr DeploymentRestriction) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":    dr.Id,
		"mode":  dr.Mode,
		"type":  dr.Type,
		"value": dr.Value,
	}
	terraformObjectValue := types.ObjectValueMust(deploymentRestrictionsAttrTypes, attributes)

	return terraformObjectValue
}
