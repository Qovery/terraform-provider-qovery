package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Database struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	Mode          types.String `tfsdk:"mode"`
	Accessibility types.String `tfsdk:"accessibility"`
	CPU           types.Int64  `tfsdk:"cpu"`
	Memory        types.Int64  `tfsdk:"memory"`
	Storage       types.Int64  `tfsdk:"storage"`
}

func (d Database) toCreateDatabaseRequest() qovery.DatabaseRequest {
	return qovery.DatabaseRequest{
		Name:          toString(d.Name),
		Type:          toString(d.Type),
		Version:       toString(d.Version),
		Mode:          toString(d.Mode),
		Accessibility: toStringPointer(d.Accessibility),
		Cpu:           toInt32Pointer(d.CPU),
		Memory:        toInt32Pointer(d.Memory),
		Storage:       toInt32Pointer(d.Storage),
	}
}

func (d Database) toUpdateDatabaseRequest() qovery.DatabaseEditRequest {
	return qovery.DatabaseEditRequest{
		Name:          toStringPointer(d.Name),
		Version:       toStringPointer(d.Version),
		Accessibility: toStringPointer(d.Accessibility),
		Cpu:           toInt32Pointer(d.CPU),
		Memory:        toInt32Pointer(d.Memory),
		Storage:       toInt32Pointer(d.Storage),
	}
}

func convertResponseToDatabase(database *qovery.DatabaseResponse) Database {
	return Database{
		Id:            fromString(database.Id),
		EnvironmentId: fromString(database.Environment.Id),
		Name:          fromString(database.Name),
		Type:          fromString(database.Type),
		Version:       fromString(database.Version),
		Mode:          fromString(database.Mode),
		Accessibility: fromStringPointer(database.Accessibility),
		CPU:           fromInt32Pointer(database.Cpu),
		Memory:        fromInt32Pointer(database.Memory),
		Storage:       fromInt32Pointer(database.Storage),
	}
}
