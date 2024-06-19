package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/qovery/qovery-client-go"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
)

type LabelsGroup struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Labels         types.Set    `tfsdk:"labels"`
}

type LabelDomain struct {
	Key                      types.String `tfsdk:"key"`
	Value                    types.String `tfsdk:"value"`
	PropagateToCloudProvider types.Bool   `tfsdk:"propagate_to_cloud_provider"`
}

type LabelList []LabelDomain

func (lg LabelsGroup) toUpsertRequest() (*labels_group.UpsertServiceRequest, error) {
	labels := make([]labels_group.LabelUpsertRequest, 0, len(lg.Labels.Elements()))
	for _, elem := range lg.Labels.Elements() {
		labels = append(labels, toLabel(elem.(types.Object)))
	}

	return &labels_group.UpsertServiceRequest{
		LabelsGroupUpsertRequest: labels_group.UpsertRequest{
			Name:   ToString(lg.Name),
			Labels: labels,
		},
	}, nil
}

func toLabel(v types.Object) labels_group.LabelUpsertRequest {
	return labels_group.LabelUpsertRequest{
		Key:                      v.Attributes()["key"].(types.String).ValueString(),
		Value:                    v.Attributes()["value"].(types.String).ValueString(),
		PropagateToCloudProvider: v.Attributes()["propagate_to_cloud_provider"].(types.Bool).ValueBool(),
	}
}

//	func (label LabelDomain) toTerraformObject() attr.Value {
//		var attributes = map[string]attr.Value{
//			"key":   label.Key,
//			"value": label.Value,
//		}
//		terraformObjectValue, diagnostics := types.ObjectValue(labelAttrTypes, attributes)
//		if diagnostics.HasError() {
//			panic("Can't creat e ObjectValue")
//		}
//		return terraformObjectValue
//	}
//
//	var labelAttrTypes = map[string]attr.Type{
//		"key":   types.StringType,
//		"value": types.StringType,
//	}

var labelsGroupAttrTypes = map[string]attr.Type{
	"key":                         types.StringType,
	"value":                       types.StringType,
	"propagate_to_cloud_provider": types.BoolType,
}

func (l LabelDomain) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"key":                         l.Key,
		"value":                       l.Value,
		"propagate_to_cloud_provider": l.PropagateToCloudProvider,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(labelsGroupAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (labels LabelList) toTerraformSet(ctx context.Context) types.Set {
	var labelGroupObjectType = types.ObjectType{
		AttrTypes: labelsGroupAttrTypes,
	}
	if labels == nil {
		return types.SetNull(labelGroupObjectType)
	}

	var elements = make([]attr.Value, 0, len(labels))
	for _, label := range labels {
		elements = append(elements, label.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, labelGroupObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func convertResponseToLabelsGroup(ctx context.Context, state LabelsGroup, labelsGroup *labels_group.LabelsGroup) LabelsGroup {
	return LabelsGroup{
		Id:             FromString(labelsGroup.Id.String()),
		Name:           FromString(labelsGroup.Name),
		OrganizationId: FromString(state.OrganizationId.ValueString()),
		Labels:         fromLabelList(labelsGroup.Labels).toTerraformSet(ctx),
	}
}

func fromLabel(label qovery.Label) LabelDomain {
	return LabelDomain{
		Key:                      FromString(label.Key),
		Value:                    FromString(label.Value),
		PropagateToCloudProvider: FromBool(label.PropagateToCloudProvider),
	}
}

func fromLabelList(labels []qovery.Label) LabelList {
	list := make([]LabelDomain, 0, len(labels))
	for _, label := range labels {
		list = append(list, fromLabel(label))
	}

	return list
}

func fromLabelsGroupResponseList(ctx context.Context, initialState types.Set, labelsGroup []qovery.OrganizationLabelsGroupResponse) types.Set {
	if initialState.IsNull() {
		return types.SetNull(types.StringType)
	}

	var elements = make([]string, 0, len(labelsGroup))
	for _, v := range labelsGroup {
		elements = append(elements, v.Id)
	}
	set, diagnostics := types.SetValueFrom(ctx, types.StringType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func fromLabelsGroupList(ctx context.Context, initialState types.Set, labelsGroup []string) types.Set {
	if initialState.IsNull() {
		return types.SetNull(types.StringType)
	}

	var elements = make([]string, 0, len(labelsGroup))
	for _, v := range labelsGroup {
		elements = append(elements, v)
	}
	set, diagnostics := types.SetValueFrom(ctx, types.StringType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}
