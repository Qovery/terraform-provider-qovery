package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"sort"
	"strings"
)

var portAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"protocol":            types.StringType,
	"internal_port":       types.Int64Type,
	"external_port":       types.Int64Type,
	"publicly_accessible": types.BoolType,
	"is_default":          types.BoolType,
}

type PortList []Port

func (pp PortList) toTerraformList(ctx context.Context) types.List {
	var portObjectType = types.ObjectType{
		AttrTypes: portAttrTypes,
	}
	if pp == nil {
		return types.ListNull(portObjectType)
	}

	sort.Slice(pp, func(i, j int) bool {
		return strings.Compare(pp[i].Name.String(), pp[j].Name.String()) < 0
	})

	var elements = make([]attr.Value, 0, len(pp))
	for _, v := range pp {
		elements = append(elements, v.toTerraformObject())
	}
	list, diagnostics := types.ListValueFrom(ctx, portObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}

	return list
}

type Port struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	InternalPort       types.Int64  `tfsdk:"internal_port"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	PubliclyAccessible types.Bool   `tfsdk:"publicly_accessible"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
}

func (p Port) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"id":                  p.Id,
		"name":                p.Name,
		"protocol":            p.Protocol,
		"internal_port":       p.InternalPort,
		"external_port":       p.ExternalPort,
		"publicly_accessible": p.PubliclyAccessible,
		"is_default":          p.IsDefault,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(portAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (p Port) toUpsertRequest() port.UpsertRequest {
	return port.UpsertRequest{
		Id:                 ToStringPointer(p.Id),
		Name:               ToStringPointer(p.Name),
		Protocol:           ToStringPointer(p.Protocol),
		InternalPort:       ToInt32(p.InternalPort),
		ExternalPort:       ToInt32Pointer(p.ExternalPort),
		PubliclyAccessible: ToBool(p.PubliclyAccessible),
		IsDefault:          ToBool(p.IsDefault),
	}
}

func fromPort(p port.Port) Port {
	return Port{
		Id:                 FromString(p.ID.String()),
		Name:               FromStringPointer(p.Name),
		Protocol:           FromString(p.Protocol.String()),
		InternalPort:       FromInt32(p.InternalPort),
		ExternalPort:       FromInt32Pointer(p.ExternalPort),
		PubliclyAccessible: FromBool(p.PubliclyAccessible),
		IsDefault:          FromBool(p.IsDefault),
	}
}

func fromPortList(state PortList, ports port.Ports) PortList {
	list := make([]Port, 0, len(ports))
	for _, s := range ports {
		list = append(list, fromPort(s))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func convertDomainPortsToPortList(initialState types.List, ports port.Ports) PortList {
	list := make([]Port, 0, len(ports))
	for _, s := range ports {
		list = append(list, convertDomainPortToPort(s))
	}

	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	return list
}

func convertDomainPortToPort(s port.Port) Port {
	return Port{
		Id:                 FromString(s.ID.String()),
		Name:               FromStringPointer(s.Name),
		Protocol:           FromString(s.Protocol.String()),
		InternalPort:       FromInt32(s.InternalPort),
		ExternalPort:       FromInt32Pointer(s.ExternalPort),
		PubliclyAccessible: FromBool(s.PubliclyAccessible),
		IsDefault:          FromBool(s.IsDefault),
	}
}

func toPort(v types.Object) Port {
	return Port{
		Id:                 v.Attributes()["id"].(types.String),
		Name:               v.Attributes()["name"].(types.String),
		Protocol:           v.Attributes()["protocol"].(types.String),
		InternalPort:       v.Attributes()["internal_port"].(types.Int64),
		ExternalPort:       v.Attributes()["external_port"].(types.Int64),
		PubliclyAccessible: v.Attributes()["publicly_accessible"].(types.Bool),
		IsDefault:          v.Attributes()["is_default"].(types.Bool),
	}
}

func toPortList(vars types.List) PortList {
	if vars.IsNull() || vars.IsUnknown() {
		return []Port{}
	}

	ports := make([]Port, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		ports = append(ports, toPort(elem.(types.Object)))
	}

	return ports
}
