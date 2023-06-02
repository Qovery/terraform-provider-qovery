package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

var portAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"protocol":            types.StringType,
	"internal_port":       types.Int64Type,
	"external_port":       types.Int64Type,
	"publicly_accessible": types.BoolType,
	"is_default":          types.BoolType,
	"has_readiness_probe": types.BoolType,
	"has_liveness_probe":  types.BoolType,
}

type PortList []Port

func (pp PortList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: portAttrTypes,
		},
	}

	if pp == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(pp))
	for _, v := range pp {
		set.Elems = append(set.Elems, v.toTerraformObject())
	}
	return set
}

type Port struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	InternalPort       types.Int64  `tfsdk:"internal_port"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	PubliclyAccessible types.Bool   `tfsdk:"publicly_accessible"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	HasReadinessProbe  types.Bool   `tfsdk:"has_readiness_probe"`
	HasLivenessProbe   types.Bool   `tfsdk:"has_liveness_probe"`
}

func (p Port) toTerraformObject() types.Object {
	return types.Object{
		AttrTypes: portAttrTypes,
		Attrs: map[string]attr.Value{
			"id":                  p.Id,
			"name":                p.Name,
			"protocol":            p.Protocol,
			"internal_port":       p.InternalPort,
			"external_port":       p.ExternalPort,
			"publicly_accessible": p.PubliclyAccessible,
			"is_default":          p.IsDefault,
			"has_readiness_probe": p.HasReadinessProbe,
			"has_liveness_probe":  p.HasLivenessProbe,
		},
	}
}

func (p Port) toUpsertRequest() port.UpsertRequest {
	return port.UpsertRequest{
		Name:               ToStringPointer(p.Name),
		Protocol:           ToStringPointer(p.Protocol),
		InternalPort:       ToInt32(p.InternalPort),
		ExternalPort:       ToInt32Pointer(p.ExternalPort),
		PubliclyAccessible: ToBool(p.PubliclyAccessible),
		IsDefault:          ToBool(p.IsDefault),
		HasReadinessProbe:  ToBool(p.HasReadinessProbe),
		HasLivenessProbe:   ToBool(p.HasLivenessProbe),
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
		HasLivenessProbe:   FromBool(p.HasReadinessProbe),
		HasReadinessProbe:  FromBool(p.HasLivenessProbe),
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

func convertDomainPortsToPortList(ports port.Ports) PortList {
	list := make([]Port, 0, len(ports))
	for _, s := range ports {
		list = append(list, convertDomainPortToPort(s))
	}

	if len(list) == 0 {
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
		HasLivenessProbe:   FromBool(s.HasReadinessProbe),
		HasReadinessProbe:  FromBool(s.HasLivenessProbe),
	}
}

func toPort(v types.Object) Port {
	return Port{
		Id:                 v.Attrs["id"].(types.String),
		Name:               v.Attrs["name"].(types.String),
		Protocol:           v.Attrs["protocol"].(types.String),
		InternalPort:       v.Attrs["internal_port"].(types.Int64),
		ExternalPort:       v.Attrs["external_port"].(types.Int64),
		PubliclyAccessible: v.Attrs["publicly_accessible"].(types.Bool),
		IsDefault:          v.Attrs["is_default"].(types.Bool),
		HasLivenessProbe:   v.Attrs["has_readiness_probe"].(types.Bool),
		HasReadinessProbe:  v.Attrs["has_liveness_probe"].(types.Bool),
	}
}

func toPortList(vars types.Set) PortList {
	if vars.Null || vars.Unknown {
		return []Port{}
	}

	ports := make([]Port, 0, len(vars.Elems))
	for _, elem := range vars.Elems {
		ports = append(ports, toPort(elem.(types.Object)))
	}

	return ports
}
