package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
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
	State         types.String `tfsdk:"state"`
}

func (d Database) toCreateDatabaseRequest() client.DatabaseCreateParams {
	return client.DatabaseCreateParams{
		DatabaseRequest: qovery.DatabaseRequest{
			Name:          toString(d.Name),
			Type:          toString(d.Type),
			Version:       toString(d.Version),
			Mode:          toString(d.Mode),
			Accessibility: toStringPointer(d.Accessibility),
			Cpu:           toInt32Pointer(d.CPU),
			Memory:        toInt32Pointer(d.Memory),
			Storage:       toInt32Pointer(d.Storage),
		},
		DesiredState: d.State.Value,
	}
}

func (d Database) toUpdateDatabaseRequest() client.DatabaseUpdateParams {
	return client.DatabaseUpdateParams{
		DatabaseEditRequest: qovery.DatabaseEditRequest{
			Name:          toStringPointer(d.Name),
			Version:       toStringPointer(d.Version),
			Accessibility: toStringPointer(d.Accessibility),
			Cpu:           toInt32Pointer(d.CPU),
			Memory:        toInt32Pointer(d.Memory),
			Storage:       toInt32Pointer(d.Storage),
		},
		DesiredState: d.State.Value,
	}
}

func convertResponseToDatabase(res *client.DatabaseResponse) Database {
	return Database{
		Id:            fromString(res.DatabaseResponse.Id),
		EnvironmentId: fromString(res.DatabaseResponse.Environment.Id),
		Name:          fromString(res.DatabaseResponse.Name),
		Type:          fromString(res.DatabaseResponse.Type),
		Version:       fromString(res.DatabaseResponse.Version),
		Mode:          fromString(res.DatabaseResponse.Mode),
		Accessibility: fromStringPointer(res.DatabaseResponse.Accessibility),
		CPU:           fromInt32Pointer(res.DatabaseResponse.Cpu),
		Memory:        fromInt32Pointer(res.DatabaseResponse.Memory),
		Storage:       fromInt32Pointer(res.DatabaseResponse.Storage),
		State:         fromString(res.DatabaseStatus.State),
	}
}
