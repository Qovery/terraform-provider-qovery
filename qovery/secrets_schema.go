package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// secretsResourceAttribute is the secrets attribute of a resource. It is a list, not a set: the
// framework refuses a write-only attribute in a set, whose elements are identified by their whole
// value. The entries are matched by key, so moving one plans no API call.
//
// The whole list is sensitive, as the set was: a set cannot mark part of an element, so Terraform
// marked every secrets set of an existing state sensitive. With the value alone sensitive, the
// marks would differ and the first plan after the upgrade would update every resource that has
// secrets.
//
// The id has no UseStateForUnknown: in a list it would copy the id of the entry that held the
// same index, so a move would plan a wrong id. Any change plans it unknown, and the read sets it.
func secretsResourceAttribute(d variableListDescription) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: d.List + " Each entry sets `value` or `value_wo`, and each key appears once.",
		Optional:            true,
		Sensitive:           true,
		Validators:          []validator.List{secretsValidator{}},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: d.ID,
					Computed:            true,
				},
				"key": schema.StringAttribute{
					MarkdownDescription: d.Key,
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: d.Value + " Set it or `value_wo`.",
					Optional:            true,
					Sensitive:           true,
				},
				"value_wo": schema.StringAttribute{
					MarkdownDescription: writeOnlyDescription("value"),
					Optional:            true,
					Sensitive:           true,
					WriteOnly:           true,
				},
				"value_wo_version": schema.StringAttribute{
					MarkdownDescription: writeOnlyVersionDescription("value", "as does a change of `description`"),
					Optional:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: d.Description,
					Optional:            true,
				},
			},
		},
	}
}

// secretsValidator checks each secrets entry: it sets value or value_wo, the write-only rules of
// validateWriteOnly hold, and no other entry has its key. A list does not merge identical entries
// as a set did, and the API keeps one secret per key. A check waits until the values it reads are
// known.
type secretsValidator struct{}

func (v secretsValidator) Description(_ context.Context) string {
	return "each entry sets value or value_wo, and each key appears once"
}

func (v secretsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v secretsValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	firstIndexByKey := map[string]int{}
	for i, element := range req.ConfigValue.Elements() {
		object, ok := element.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			continue
		}
		entry := toSecret(object)
		entryPath := req.Path.AtListIndex(i)

		resp.Diagnostics.Append(validateWriteOnly(entryPath, "value", entry.Value, entry.ValueWo, entry.ValueWoVersion)...)
		if entry.Value.IsNull() && entry.ValueWo.IsNull() {
			resp.Diagnostics.AddAttributeError(
				entryPath.AtName("value"),
				"Missing secret value",
				fmt.Sprintf("%s or %s is required.", entryPath.AtName("value"), entryPath.AtName("value_wo")),
			)
		}

		if !isKnownAndSet(entry.Key) {
			continue
		}
		if first, ok := firstIndexByKey[entry.Key.ValueString()]; ok {
			resp.Diagnostics.AddAttributeError(
				entryPath.AtName("key"),
				"Duplicate secret key",
				fmt.Sprintf("%s is the key of %s too: each key can appear once.", entryPath.AtName("key"), req.Path.AtListIndex(first)),
			)
			continue
		}
		firstIndexByKey[entry.Key.ValueString()] = i
	}
}
