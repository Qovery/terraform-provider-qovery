package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Database struct {
	Id                types.String `tfsdk:"id"`
	EnvironmentId     types.String `tfsdk:"environment_id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Version           types.String `tfsdk:"version"`
	Mode              types.String `tfsdk:"mode"`
	Accessibility     types.String `tfsdk:"accessibility"`
	CPU               types.Int64  `tfsdk:"cpu"`
	Memory            types.Int64  `tfsdk:"memory"`
	ExternalHost      types.String `tfsdk:"external_host"`
	InternalHost      types.String `tfsdk:"internal_host"`
	Port              types.Int64  `tfsdk:"port"`
	Login             types.String `tfsdk:"login"`
	Password          types.String `tfsdk:"password"`
	Storage           types.Int64  `tfsdk:"storage"`
	DeploymentStageId types.String `tfsdk:"deployment_stage_id"`
}

func (d Database) toCreateDatabaseRequest() (*client.DatabaseCreateParams, error) {
	dbType, err := qovery.NewDatabaseTypeEnumFromValue(toString(d.Type))
	if err != nil {
		return nil, err
	}

	mode, err := qovery.NewDatabaseModeEnumFromValue(toString(d.Mode))
	if err != nil {
		return nil, err
	}

	accessibility, err := qovery.NewDatabaseAccessibilityEnumFromValue(toString(d.Accessibility))
	if err != nil {
		return nil, err
	}

	return &client.DatabaseCreateParams{
		DatabaseRequest: qovery.DatabaseRequest{
			Name:          toString(d.Name),
			Type:          *dbType,
			Version:       toString(d.Version),
			Mode:          *mode,
			Accessibility: accessibility,
			Cpu:           toInt32Pointer(d.CPU),
			Memory:        toInt32Pointer(d.Memory),
			Storage:       toInt32Pointer(d.Storage),
		},
		DeploymentStageId: toString(d.DeploymentStageId),
	}, nil
}

func (d Database) toUpdateDatabaseRequest() (*client.DatabaseUpdateParams, error) {
	accessibility, err := qovery.NewDatabaseAccessibilityEnumFromValue(toString(d.Accessibility))
	if err != nil {
		return nil, err
	}

	return &client.DatabaseUpdateParams{
		DatabaseEditRequest: qovery.DatabaseEditRequest{
			Name:          toStringPointer(d.Name),
			Version:       toStringPointer(d.Version),
			Accessibility: accessibility,
			Cpu:           toInt32Pointer(d.CPU),
			Memory:        toInt32Pointer(d.Memory),
			Storage:       toInt32Pointer(d.Storage),
		},
	}, nil
}

func convertResponseToDatabase(res *client.DatabaseResponse) Database {
	return Database{
		Id:                fromString(res.DatabaseResponse.Id),
		EnvironmentId:     fromString(res.DatabaseResponse.Environment.Id),
		Name:              fromString(res.DatabaseResponse.Name),
		Type:              fromClientEnum(res.DatabaseResponse.Type),
		Version:           fromString(res.DatabaseResponse.Version),
		Mode:              fromClientEnum(res.DatabaseResponse.Mode),
		Accessibility:     fromClientEnumPointer(res.DatabaseResponse.Accessibility),
		CPU:               fromInt32Pointer(res.DatabaseResponse.Cpu),
		Memory:            fromInt32Pointer(res.DatabaseResponse.Memory),
		ExternalHost:      fromString(res.DatabaseResponse.GetHost()),
		InternalHost:      fromString(res.DatabaseInternalHost),
		Port:              fromInt32Pointer(res.DatabaseResponse.Port),
		Login:             fromString(res.DatabaseCredentials.Login),
		Password:          fromString(res.DatabaseCredentials.Password),
		Storage:           fromInt32Pointer(res.DatabaseResponse.Storage),
		DeploymentStageId: fromString(res.DeploymentStageId),
	}
}
