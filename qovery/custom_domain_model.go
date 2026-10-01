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
	"use_cdn":              types.BoolType,
}

type CustomDomainList []CustomDomain

func (domains CustomDomainList) toTerraformSet(ctx context.Context) types.Set {
	domainObjectType := types.ObjectType{
		AttrTypes: customDomainAttrTypes,
	}
	if domains == nil {
		return types.SetNull(domainObjectType)
	}

	elements := make([]attr.Value, 0, len(domains))
	for _, d := range domains {
		elements = append(elements, d.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, domainObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}

	return set
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

	for _, domain := range domains {
		oldDomain := oldDomains.find(ToString(domain.Domain))
		if oldDomain == nil {
			diff.Create = append(diff.Create, domain.toCreateRequest())
		} else if oldDomain.GenerateCertificate != domain.GenerateCertificate || oldDomain.UseCdn != domain.UseCdn {
			diff.Update = append(diff.Update, oldDomain.toUpdateRequest(domain))
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
	UseCdn              types.Bool   `tfsdk:"use_cdn"`
}

func (d CustomDomain) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":                   d.Id,
		"domain":               d.Domain,
		"validation_domain":    d.ValidationDomain,
		"status":               d.Status,
		"generate_certificate": d.GenerateCertificate,
		"use_cdn":              d.UseCdn,
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
			UseCdn:              ToBoolPointer(d.UseCdn),
		},
	}
}

func (d CustomDomain) toUpdateRequest(new CustomDomain) client.CustomDomainUpdateRequest {
	return client.CustomDomainUpdateRequest{
		Id: ToString(d.Id),
		CustomDomainRequest: qovery.CustomDomainRequest{
			Domain:              ToString(new.Domain),
			GenerateCertificate: ToBool(new.GenerateCertificate),
			UseCdn:              ToBoolPointer(new.UseCdn),
		},
	}
}

func (d CustomDomain) toDeleteRequest() client.CustomDomainDeleteRequest {
	return client.CustomDomainDeleteRequest{
		Id: ToString(d.Id),
	}
}

// fromCustomDomain converts an API custom domain. generate_certificate and use_cdn report the
// API value, so a change made in the Console shows up in the plan. q-core stores use_cdn false
// when a request omits it, which is also the value the schema plans when it is omitted.
func fromCustomDomain(domain *qovery.CustomDomain) CustomDomain {
	useCdn := false
	if domain.UseCdn != nil {
		useCdn = *domain.UseCdn
	}

	return CustomDomain{
		Id:                  FromString(domain.Id),
		Domain:              FromString(domain.Domain),
		ValidationDomain:    FromStringPointer(domain.ValidationDomain),
		Status:              fromClientEnumPointer(domain.Status),
		GenerateCertificate: FromBool(domain.GenerateCertificate),
		UseCdn:              FromBool(useCdn),
	}
}

// fromCustomDomainList converts the API custom domains. No domain keeps the shape of
// initialState: null stays null and [] stays [].
func fromCustomDomainList(initialState types.Set, customDomains []*qovery.CustomDomain) CustomDomainList {
	list := make([]CustomDomain, 0, len(customDomains))
	for _, customDomain := range customDomains {
		list = append(list, fromCustomDomain(customDomain))
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
		UseCdn:              v.Attributes()["use_cdn"].(types.Bool),
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
