package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/qoveryapi"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Database struct {
	Id                  types.String `tfsdk:"id"`
	EnvironmentId       types.String `tfsdk:"environment_id"`
	Name                types.String `tfsdk:"name"`
	IconUri             types.String `tfsdk:"icon_uri"`
	Type                types.String `tfsdk:"type"`
	Version             types.String `tfsdk:"version"`
	Mode                types.String `tfsdk:"mode"`
	Accessibility       types.String `tfsdk:"accessibility"`
	CPU                 types.Int64  `tfsdk:"cpu"`
	Memory              types.Int64  `tfsdk:"memory"`
	ExternalHost        types.String `tfsdk:"external_host"`
	InternalHost        types.String `tfsdk:"internal_host"`
	Port                types.Int64  `tfsdk:"port"`
	Login               types.String `tfsdk:"login"`
	Password            types.String `tfsdk:"password"`
	Storage             types.Int64  `tfsdk:"storage"`
	InstanceType        types.String `tfsdk:"instance_type"`
	DeploymentStageId   types.String `tfsdk:"deployment_stage_id"`
	IsSkipped           types.Bool   `tfsdk:"is_skipped"`
	AnnotationsGroupIds types.Set    `tfsdk:"annotations_group_ids"`
	LabelsGroupIds      types.Set    `tfsdk:"labels_group_ids"`
}

func (d Database) toCreateDatabaseRequest() (*client.DatabaseCreateParams, error) {
	dbType, err := qovery.NewDatabaseTypeEnumFromValue(ToString(d.Type))
	if err != nil {
		return nil, err
	}

	mode, err := qovery.NewDatabaseModeEnumFromValue(ToString(d.Mode))
	if err != nil {
		return nil, err
	}

	accessibility, err := qovery.NewDatabaseAccessibilityEnumFromValue(ToString(d.Accessibility))
	if err != nil {
		return nil, err
	}

	annotations := make([]string, 0, len(d.AnnotationsGroupIds.Elements()))
	for _, id := range d.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotations = append(annotations, id.ValueString())
	}

	annotationsGroups, err := qoveryapi.NewQoveryServiceAnnotationsGroupRequestFromDomain(annotations)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labels := make([]string, 0, len(d.LabelsGroupIds.Elements()))
	for _, id := range d.LabelsGroupIds.Elements() {
		id := id.(types.String)
		labels = append(labels, id.ValueString())
	}

	labelsGroups, err := qoveryapi.NewQoveryServiceLabelsGroupRequestFromDomain(labels)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	return &client.DatabaseCreateParams{
		DatabaseRequest: qovery.DatabaseRequest{
			Name:              ToString(d.Name),
			IconUri:           ToStringPointer(d.IconUri),
			Type:              *dbType,
			Version:           ToString(d.Version),
			Mode:              *mode,
			Accessibility:     accessibility,
			Cpu:               ToInt32Pointer(d.CPU),
			Memory:            ToInt32Pointer(d.Memory),
			Storage:           ToInt32Pointer(d.Storage),
			InstanceType:      ToStringPointer(d.InstanceType),
			AnnotationsGroups: annotationsGroups,
			LabelsGroups:      labelsGroups,
		},
		DeploymentStageID: ToString(d.DeploymentStageId),
		IsSkipped:         ToBool(d.IsSkipped),
	}, nil
}

func (d Database) toUpdateDatabaseRequest() (*client.DatabaseUpdateParams, error) {
	accessibility, err := qovery.NewDatabaseAccessibilityEnumFromValue(ToString(d.Accessibility))
	if err != nil {
		return nil, err
	}

	annotations := make([]string, 0, len(d.AnnotationsGroupIds.Elements()))
	for _, id := range d.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotations = append(annotations, id.ValueString())
	}

	annotationsGroups, err := qoveryapi.NewQoveryServiceAnnotationsGroupRequestFromDomain(annotations)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labels := make([]string, 0, len(d.LabelsGroupIds.Elements()))
	for _, id := range d.LabelsGroupIds.Elements() {
		id := id.(types.String)
		labels = append(labels, id.ValueString())
	}

	labelsGroups, err := qoveryapi.NewQoveryServiceLabelsGroupRequestFromDomain(labels)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	return &client.DatabaseUpdateParams{
		DatabaseEditRequest: qovery.DatabaseEditRequest{
			Name:              ToStringPointer(d.Name),
			IconUri:           ToStringPointer(d.IconUri),
			Version:           ToStringPointer(d.Version),
			Accessibility:     accessibility,
			Cpu:               ToInt32Pointer(d.CPU),
			Memory:            ToInt32Pointer(d.Memory),
			Storage:           ToInt32Pointer(d.Storage),
			InstanceType:      ToStringPointer(d.InstanceType),
			AnnotationsGroups: annotationsGroups,
			LabelsGroups:      labelsGroups,
		},
		IsSkipped: ToBool(d.IsSkipped),
	}, nil
}

func convertResponseToDatabase(ctx context.Context, state Database, res *client.DatabaseResponse) Database {

	return Database{
		Id:                  FromString(res.DatabaseResponse.Id),
		EnvironmentId:       FromString(res.DatabaseResponse.Environment.Id),
		Name:                FromString(res.DatabaseResponse.Name),
		IconUri:             FromString(res.DatabaseResponse.IconUri),
		Type:                fromClientEnum(res.DatabaseResponse.Type),
		Version:             FromString(res.DatabaseResponse.Version),
		Mode:                fromClientEnum(res.DatabaseResponse.Mode),
		Accessibility:       fromClientEnumPointer(res.DatabaseResponse.Accessibility),
		CPU:                 FromInt32Pointer(res.DatabaseResponse.Cpu),
		Memory:              FromInt32Pointer(res.DatabaseResponse.Memory),
		ExternalHost:        FromString(res.DatabaseResponse.GetHost()),
		InternalHost:        FromString(res.DatabaseInternalHost),
		Port:                FromInt32Pointer(res.DatabaseResponse.Port),
		Login:               FromString(res.DatabaseCredentials.Login),
		Password:            FromString(res.DatabaseCredentials.Password),
		Storage:             FromInt32Pointer(res.DatabaseResponse.Storage),
		DeploymentStageId:   FromString(res.DeploymentStageID),
		IsSkipped:           FromBool(res.IsSkipped),
		InstanceType:        FromStringPointer(res.DatabaseResponse.InstanceType),
		AnnotationsGroupIds: fromAnnotationsGroupResponseList(ctx, state.AnnotationsGroupIds, res.DatabaseResponse.AnnotationsGroups),
		LabelsGroupIds:      fromLabelsGroupResponseList(ctx, state.LabelsGroupIds, res.DatabaseResponse.LabelsGroups),
	}
}
