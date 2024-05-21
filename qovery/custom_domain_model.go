package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

var customDomainAttrTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"domain":               types.StringType,
	"validation_domain":    types.StringType,
	"status":               types.StringType,
	"generate_certificate": types.BoolType,
}

type CustomDomainList []CustomDomain

func (domains CustomDomainList) toTerraformSet(ctx context.Context) types.Set {
	var domainObjectType = types.ObjectType{
		AttrTypes: customDomainAttrTypes,
	}
	if domains == nil {
		return types.SetNull(domainObjectType)
	}

	var elements = make([]attr.Value, 0, len(domains))
	for _, d := range domains {
		elements = append(elements, d.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, domainObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}

	return set
}

func (domains CustomDomainList) contains(domain CustomDomain) bool {
	for _, d := range domains {
		if domain.Domain == d.Domain {
			return true
		}
	}
	return false
}

func (domains CustomDomainList) find(domain string) *CustomDomain {
	for _, d := range domains {
		if d.Domain.ValueString() == domain {
			return &d
		}
	}
	return nil
}

func (domains CustomDomainList) diff(oldDomains CustomDomainList) client.CustomDomainsDiff {
	diff := client.CustomDomainsDiff{
		Create: []client.CustomDomainCreateRequest{},
		Update: []client.CustomDomainUpdateRequest{},
		Delete: []client.CustomDomainDeleteRequest{},
	}

	for _, od := range oldDomains {
		if found := domains.find(ToString(od.Domain)); found == nil {
			diff.Delete = append(diff.Delete, od.toDeleteRequest())
		}
	}

	for _, d := range domains {
		do := oldDomains.find(ToString(d.Domain))
		if do == nil {
			diff.Create = append(diff.Create, d.toCreateRequest())
		} else if do.GenerateCertificate != d.GenerateCertificate {
			diff.Update = append(diff.Update, do.toUpdateRequest(d))
		}
	}

	return diff
}

type CustomDomain struct {
	Id                  types.String `tfsdk:"id"`
	Domain              types.String `tfsdk:"domain"`
	ValidationDomain    types.String `tfsdk:"validation_domain"`
	Status              types.String `tfsdk:"status"`
	GenerateCertificate types.Bool   `tfsdk:"generate_certificate"`
}

func (d CustomDomain) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"id":                   d.Id,
		"domain":               d.Domain,
		"validation_domain":    d.ValidationDomain,
		"status":               d.Status,
		"generate_certificate": d.GenerateCertificate,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(customDomainAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (d CustomDomain) toCreateRequest() client.CustomDomainCreateRequest {
	return client.CustomDomainCreateRequest{
		CustomDomainRequest: qovery.CustomDomainRequest{
			Domain:              ToString(d.Domain),
			GenerateCertificate: ToBool(d.GenerateCertificate),
		},
	}
}

func (d CustomDomain) toUpdateRequest(new CustomDomain) client.CustomDomainUpdateRequest {
	return client.CustomDomainUpdateRequest{
		Id: ToString(d.Id),
		CustomDomainRequest: qovery.CustomDomainRequest{
			Domain:              ToString(new.Domain),
			GenerateCertificate: ToBool(new.GenerateCertificate),
		},
	}
}

func (d CustomDomain) toDeleteRequest() client.CustomDomainDeleteRequest {
	return client.CustomDomainDeleteRequest{
		Id: ToString(d.Id),
	}
}

func fromCustomDomain(plan *CustomDomain, d *qovery.CustomDomain) CustomDomain {
	var generateCertificate *bool
	if plan != nil && (plan.GenerateCertificate.IsNull() || plan.GenerateCertificate.IsUnknown()) {
		// as GenerateCertificate is optional, terraform expect to receive null if GenerateCertificate is not defined in the plan
		generateCertificate = nil
	} else {
		generateCertificate = &d.GenerateCertificate
	}

	return CustomDomain{
		Id:                  FromString(d.Id),
		Domain:              FromString(d.Domain),
		ValidationDomain:    FromStringPointer(d.ValidationDomain),
		Status:              fromClientEnumPointer(d.Status),
		GenerateCertificate: FromBoolPointer(generateCertificate),
	}
}

func findCustomDomainByDomain(initialState types.Set, domain string) *CustomDomain {
	for _, elem := range initialState.Elements() {
		customDomain := toCustomDomain(elem.(types.Object))
		if customDomain.Domain.ValueString() == domain {
			return &customDomain
		}
	}
	return nil
}

func fromCustomDomainList(initialState types.Set, customDomains []*qovery.CustomDomain) CustomDomainList {
	list := make([]CustomDomain, 0, len(customDomains))
	for _, customDomain := range customDomains {
		found := findCustomDomainByDomain(initialState, customDomain.Domain)
		list = append(list, fromCustomDomain(found, customDomain))
	}

	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	return list
}

func toCustomDomain(v types.Object) CustomDomain {
	return CustomDomain{
		Id:                  v.Attributes()["id"].(types.String),
		Domain:              v.Attributes()["domain"].(types.String),
		ValidationDomain:    v.Attributes()["validation_domain"].(types.String),
		Status:              v.Attributes()["status"].(types.String),
		GenerateCertificate: v.Attributes()["generate_certificate"].(types.Bool),
	}
}

func toCustomDomainList(vars types.Set) CustomDomainList {
	if vars.IsNull() || vars.IsUnknown() {
		return nil
	}

	customDomains := make([]CustomDomain, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		customDomains = append(customDomains, toCustomDomain(elem.(types.Object)))
	}

	return customDomains
}
