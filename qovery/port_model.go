package qovery

import (
	"context"
	"sort"

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
}

type PortList []Port

func (pp PortList) toTerraformList(ctx context.Context) types.List {
	var portObjectType = types.ObjectType{
		AttrTypes: portAttrTypes,
	}
	if pp == nil {
		return types.ListNull(portObjectType)
	}

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

// convertDomainPortsToPortList preserves state ordering using ID-first matching
// with name fallback. Keep in sync with convertResponseToApplicationPorts in
// resource_application_model.go which implements the same algorithm for API types.
func convertDomainPortsToPortList(ctx context.Context, initialState types.List, ports port.Ports) PortList {
	if len(ports) == 0 && initialState.IsNull() {
		return nil
	}

	// Build lookup maps by ID (stable) and name (fallback).
	portsByID := make(map[string]port.Port, len(ports))
	portsByName := make(map[string]port.Port, len(ports))
	for _, p := range ports {
		portsByID[p.ID.String()] = p
		if p.Name != nil {
			portsByName[*p.Name] = p
		}
	}

	matched := make(map[string]bool, len(ports))
	list := make([]Port, 0, len(ports))

	// Match state ports: prefer ID match, fall back to name.
	if !initialState.IsNull() {
		initialStatePorts := make([]Port, 0, len(initialState.Elements()))
		initialState.ElementsAs(ctx, &initialStatePorts, false)
		for _, state := range initialStatePorts {
			if id := state.Id.ValueString(); id != "" {
				if p, ok := portsByID[id]; ok {
					list = append(list, convertDomainPortToPort(p))
					matched[p.ID.String()] = true
					continue
				}
			}
			if name := state.Name.ValueString(); name != "" {
				if p, ok := portsByName[name]; ok && !matched[p.ID.String()] {
					list = append(list, convertDomainPortToPort(p))
					matched[p.ID.String()] = true
				}
			}
		}
	}

	// Collect unmatched ports and sort deterministically.
	remaining := make(port.Ports, 0)
	for _, p := range ports {
		if !matched[p.ID.String()] {
			remaining = append(remaining, p)
		}
	}
	sort.Slice(remaining, func(i, j int) bool {
		if remaining[i].InternalPort != remaining[j].InternalPort {
			return remaining[i].InternalPort < remaining[j].InternalPort
		}
		return ptrStringValue(remaining[i].Name) < ptrStringValue(remaining[j].Name)
	})
	for _, p := range remaining {
		list = append(list, convertDomainPortToPort(p))
	}

	return list
}

// ptrStringValue returns the string value of a pointer, or empty string if nil.
func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
