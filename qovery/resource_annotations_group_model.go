package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
)

type AnnotationsGroup struct {
	Id             types.String      `tfsdk:"id"`
	OrganizationId types.String      `tfsdk:"organization_id"`
	Name           types.String      `tfsdk:"name"`
	Annotations    map[string]string `tfsdk:"annotations"`
	Scopes         []string          `tfsdk:"scopes"`
}

type AnnotationDomain struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type AnnotationList []AnnotationDomain

func (ag AnnotationsGroup) toUpsertRequest() *annotations_group.UpsertServiceRequest {
	annotations := make([]annotations_group.AnnotationUpsertRequest, 0, len(ag.Annotations))
	for key, value := range ag.Annotations {
		annotations = append(annotations, annotations_group.AnnotationUpsertRequest{
			Key:   key,
			Value: value,
		})
	}

	scopes := make([]string, 0, len(ag.Scopes))
	scopes = append(scopes, ag.Scopes...)

	return &annotations_group.UpsertServiceRequest{
		AnnotationsGroupUpsertRequest: annotations_group.UpsertRequest{
			Name:        ToString(ag.Name),
			Annotations: annotations,
			Scopes:      scopes,
		},
	}
}

func (annotations AnnotationList) toTerraformMap() map[string]string {
	elements := make(map[string]string, len(annotations))
	for _, annotation := range annotations {
		elements[ToString(annotation.Key)] = ToString(annotation.Value)
	}

	return elements
}

func convertResponseToAnnotationsGroup(state AnnotationsGroup, annotationsGroup *annotations_group.AnnotationsGroup) AnnotationsGroup {
	return AnnotationsGroup{
		Id:             FromString(annotationsGroup.Id.String()),
		Name:           FromString(annotationsGroup.Name),
		OrganizationId: FromString(state.OrganizationId.ValueString()),
		Annotations:    fromAnnotationList(annotationsGroup.Annotations).toTerraformMap(),
		Scopes:         fromScopeList(annotationsGroup.Scopes),
	}
}

func fromAnnotation(a qovery.Annotation) AnnotationDomain {
	return AnnotationDomain{
		Key:   FromString(a.Key),
		Value: FromString(a.Value),
	}
}

func fromAnnotationList(annotations []qovery.Annotation) AnnotationList {
	list := make([]AnnotationDomain, 0, len(annotations))
	for _, annotation := range annotations {
		list = append(list, fromAnnotation(annotation))
	}

	return list
}

func fromScopeList(scopes []qovery.OrganizationAnnotationsGroupScopeEnum) []string {
	list := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		list = append(list, fromClientEnumPointer(&scope).ValueString())
	}

	return list
}

func fromAnnotationsGroupResponseList(ctx context.Context, initialState types.Set, annotationsGroup []qovery.OrganizationAnnotationsGroupResponse) types.Set {
	if initialState.IsNull() {
		return types.SetNull(types.StringType)
	}

	elements := make([]string, 0, len(annotationsGroup))
	for _, v := range annotationsGroup {
		elements = append(elements, v.Id)
	}
	set, diagnostics := types.SetValueFrom(ctx, types.StringType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func fromAnnotationsGroupList(ctx context.Context, initialState types.Set, annotationsGroup []string) types.Set {
	if initialState.IsNull() {
		return types.SetNull(types.StringType)
	}

	elements := make([]string, 0, len(annotationsGroup))
	for _, v := range annotationsGroup {
		elements = append(elements, v)
	}
	set, diagnostics := types.SetValueFrom(ctx, types.StringType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}
