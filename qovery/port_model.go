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
	portObjectType := types.ObjectType{
		AttrTypes: portAttrTypes,
	}
	if pp == nil {
		return types.ListNull(portObjectType)
	}

	elements := make([]attr.Value, 0, len(pp))
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
	attributes := map[string]attr.Value{
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

// portIdentity holds the fields needed for port matching and ordering.
type portIdentity struct {
	id           string
	name         string
	internalPort int32
}

// reorderPortsPreservingState matches source ports to state positions using
// ID-first matching with name fallback, keeping matched ports at their
// original index positions. Unmatched ports are sorted by internal port
// then name and appended after matched entries.
func reorderPortsPreservingState[S any, O any](
	sourcePorts []S,
	identify func(S) portIdentity,
	convert func(S) O,
	stateIDs []string,
	stateNames []string,
) []O {
	// Build lookup maps by ID (stable) and name (fallback).
	portsByID := make(map[string]int, len(sourcePorts))
	portsByName := make(map[string]int, len(sourcePorts))
	for i, p := range sourcePorts {
		ident := identify(p)
		portsByID[ident.id] = i
		if ident.name != "" {
			portsByName[ident.name] = i
		}
	}

	// Match state ports at their original index positions.
	matched := make(map[int]bool, len(sourcePorts))
	type indexedMatch struct {
		sourceIdx int
		matched   bool
	}
	indexed := make([]indexedMatch, len(stateIDs))
	for i := range stateIDs {
		if stateIDs[i] != "" {
			if idx, ok := portsByID[stateIDs[i]]; ok {
				indexed[i] = indexedMatch{idx, true}
				matched[idx] = true
				continue
			}
		}
		if stateNames[i] != "" {
			if idx, ok := portsByName[stateNames[i]]; ok && !matched[idx] {
				indexed[i] = indexedMatch{idx, true}
				matched[idx] = true
			}
		}
	}

	// Collect unmatched source indices and sort deterministically.
	remaining := make([]int, 0)
	for i := range sourcePorts {
		if !matched[i] {
			remaining = append(remaining, i)
		}
	}
	sort.SliceStable(remaining, func(a, b int) bool {
		ai, bi := identify(sourcePorts[remaining[a]]), identify(sourcePorts[remaining[b]])
		if ai.internalPort != bi.internalPort {
			return ai.internalPort < bi.internalPort
		}
		return ai.name < bi.name
	})

	// Build final list: matched ports at original indices, gaps filled with remaining.
	remainIdx := 0
	list := make([]O, 0, len(sourcePorts))
	for i := range indexed {
		if indexed[i].matched {
			list = append(list, convert(sourcePorts[indexed[i].sourceIdx]))
		} else if remainIdx < len(remaining) {
			list = append(list, convert(sourcePorts[remaining[remainIdx]]))
			remainIdx++
		}
	}
	for remainIdx < len(remaining) {
		list = append(list, convert(sourcePorts[remaining[remainIdx]]))
		remainIdx++
	}

	return list
}

// convertDomainPortsToPortList preserves state ordering using ID-first matching
// with name fallback, keeping matched ports at their original index positions.
func convertDomainPortsToPortList(ctx context.Context, initialState types.List, ports port.Ports) PortList {
	if len(ports) == 0 && initialState.IsNull() {
		return nil
	}

	stateIDs, stateNames := extractPortStateIdentifiers(ctx, initialState)
	return reorderPortsPreservingState(
		[]port.Port(ports),
		func(p port.Port) portIdentity {
			return portIdentity{p.ID.String(), ptrStringValue(p.Name), p.InternalPort}
		},
		convertDomainPortToPort,
		stateIDs, stateNames,
	)
}

// extractPortStateIdentifiers extracts IDs and names from a Terraform state port list.
func extractPortStateIdentifiers(ctx context.Context, state types.List) (ids, names []string) {
	if state.IsNull() {
		return nil, nil
	}
	statePorts := make([]Port, 0, len(state.Elements()))
	state.ElementsAs(ctx, &statePorts, false)
	ids = make([]string, len(statePorts))
	names = make([]string, len(statePorts))
	for i, p := range statePorts {
		ids[i] = p.Id.ValueString()
		names[i] = p.Name.ValueString()
	}
	return ids, names
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
