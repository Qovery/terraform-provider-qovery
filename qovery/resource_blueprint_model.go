package qovery

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

const (
	defaultBlueprintIconURI = "app://qovery-console/terraform"
	// Platform-set variable, never user input
	blueprintImportIdentifierVariable = "import_identifier"
)

type Blueprint struct {
	ID              types.String            `tfsdk:"id"`
	EnvironmentID   types.String            `tfsdk:"environment_id"`
	Blueprint       types.String            `tfsdk:"blueprint"`
	Name            types.String            `tfsdk:"name"`
	Tag             types.String            `tfsdk:"tag"`
	IconURI         types.String            `tfsdk:"icon_uri"`
	Variables       types.Map               `tfsdk:"variables"`
	SecretVariables types.Map               `tfsdk:"secret_variables"`
	SpecOverrides   *BlueprintSpecOverrides `tfsdk:"spec_overrides"`
	Deploy          types.Bool              `tfsdk:"deploy"`
	ServiceID       types.String            `tfsdk:"service_id"`
	ServiceType     types.String            `tfsdk:"service_type"`
	CatalogURL      types.String            `tfsdk:"catalog_url"`
}

type BlueprintSpecOverrides struct {
	EngineVersion types.String `tfsdk:"engine_version"`
	Credentials   types.String `tfsdk:"credentials"`
	Backend       types.String `tfsdk:"backend"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	CPU           types.String `tfsdk:"cpu"`
	RAM           types.String `tfsdk:"ram"`
	Storage       types.String `tfsdk:"storage"`
}

func (o *BlueprintSpecOverrides) toDomain() *blueprint.SpecOverrides {
	if o == nil {
		return nil
	}
	return &blueprint.SpecOverrides{
		EngineVersion: ToStringPointer(o.EngineVersion),
		Credentials:   ToStringPointer(o.Credentials),
		Backend:       ToStringPointer(o.Backend),
		Timeout:       ToInt32Pointer(o.Timeout),
		CPU:           ToStringPointer(o.CPU),
		RAM:           ToStringPointer(o.RAM),
		Storage:       ToStringPointer(o.Storage),
	}
}

func (b Blueprint) toUpsertRequest(ctx context.Context) (blueprint.UpsertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	variables, d := stringMapFrom(ctx, b.Variables)
	diags.Append(d...)
	secretVariables, d := stringMapFrom(ctx, b.SecretVariables)
	diags.Append(d...)

	return blueprint.UpsertRequest{
		Name:            ToString(b.Name),
		Tag:             ToString(b.Tag),
		IconURI:         ToString(b.IconURI),
		Variables:       variables,
		SecretVariables: secretVariables,
		SpecOverrides:   b.SpecOverrides.toDomain(),
	}, diags
}

func (b Blueprint) toCreateRequest(ctx context.Context) (blueprint.CreateRequest, diag.Diagnostics) {
	upsert, diags := b.toUpsertRequest(ctx)
	return blueprint.CreateRequest{UpsertRequest: upsert, Deploy: b.Deploy.ValueBool()}, diags
}

func (b Blueprint) toUpdateRequest(ctx context.Context, state Blueprint) (blueprint.UpdateRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	upsert, d := b.toUpsertRequest(ctx)
	diags.Append(d...)
	previousVariables, d := stringMapFrom(ctx, state.Variables)
	diags.Append(d...)
	previousSecretVariables, d := stringMapFrom(ctx, state.SecretVariables)
	diags.Append(d...)

	return blueprint.UpdateRequest{
		UpsertRequest:         upsert,
		PreviousVariableNames: append(sortedKeys(previousVariables), sortedKeys(previousSecretVariables)...),
		PreviousSpecOverrides: state.SpecOverrides.toDomain(),
	}, diags
}

func (b Blueprint) configEqualIgnoringDeploy(other Blueprint) bool {
	return b.EnvironmentID.Equal(other.EnvironmentID) &&
		strings.EqualFold(b.Blueprint.ValueString(), other.Blueprint.ValueString()) &&
		b.Name.Equal(other.Name) &&
		b.Tag.Equal(other.Tag) &&
		b.IconURI.Equal(other.IconURI) &&
		b.Variables.Equal(other.Variables) &&
		b.SecretVariables.Equal(other.SecretVariables) &&
		reflect.DeepEqual(b.SpecOverrides, other.SpecOverrides)
}

// API returns no icon, spec overrides or secret values: those come from prior (plan or state).
// API also returns manifest defaults as variables, so only variables prior declares are tracked.
func convertDomainBlueprintToBlueprint(ctx context.Context, bp *blueprint.Blueprint, prior Blueprint, pendingApply bool) (Blueprint, diag.Diagnostics) {
	var diags diag.Diagnostics
	priorVariables, d := stringMapFrom(ctx, prior.Variables)
	diags.Append(d...)
	priorSecretVariables, d := stringMapFrom(ctx, prior.SecretVariables)
	diags.Append(d...)

	variables := map[string]string{}
	secretVariables := map[string]string{}
	for _, v := range bp.Variables {
		if v.IsSecret {
			if value, ok := priorSecretVariables[v.Name]; ok {
				secretVariables[v.Name] = value
			}
			continue
		}
		if v.Value == nil {
			continue
		}
		if _, declared := priorVariables[v.Name]; declared {
			variables[v.Name] = *v.Value
		}
	}

	variablesValue, d := stringMapValue(ctx, variables, prior.Variables.IsNull())
	diags.Append(d...)
	secretVariablesValue, d := stringMapValue(ctx, secretVariables, prior.SecretVariables.IsNull())
	diags.Append(d...)

	iconURI := prior.IconURI
	if iconURI.IsNull() || iconURI.IsUnknown() {
		iconURI = FromString(defaultBlueprintIconURI)
	}
	deploy := prior.Deploy
	if deploy.IsNull() || deploy.IsUnknown() {
		deploy = types.BoolValue(true)
	}

	state := Blueprint{
		ID:              FromString(bp.ID.String()),
		EnvironmentID:   FromString(bp.EnvironmentID.String()),
		Blueprint:       blueprintVersionValue(prior.Blueprint, bp.Tag),
		Name:            FromString(bp.Name),
		Tag:             FromString(bp.Tag),
		IconURI:         iconURI,
		Variables:       variablesValue,
		SecretVariables: secretVariablesValue,
		SpecOverrides:   prior.SpecOverrides,
		Deploy:          deploy,
		ServiceID:       FromStringPointer(bp.ServiceID),
		ServiceType:     FromString(string(bp.ServiceType)),
		CatalogURL:      FromString(bp.CatalogURL),
	}
	// Failed apply: keep last applied values so the next plan retries instead of going empty
	if (pendingApply || bp.LastApplyFailed()) && !prior.Tag.IsNull() && !prior.Tag.IsUnknown() {
		state.Name = prior.Name
		state.Tag = prior.Tag
		state.Variables = prior.Variables
		state.SecretVariables = prior.SecretVariables
	}
	return state, diags
}

func blueprintVersionValue(prior types.String, tag string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() {
		return prior
	}
	version, err := blueprint.CatalogVersionFromTag(tag)
	if err != nil {
		return types.StringNull()
	}
	return FromString(version.String())
}

func stringMapFrom(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	values := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return values, nil
	}
	diags := m.ElementsAs(ctx, &values, false)
	return values, diags
}

// Omitted attribute must stay null, not flip to empty map after apply
func stringMapValue(ctx context.Context, values map[string]string, nullWhenEmpty bool) (types.Map, diag.Diagnostics) {
	if len(values) == 0 && nullWhenEmpty {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
