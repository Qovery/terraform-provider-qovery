package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

var customDomainAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"domain":            types.StringType,
	"validation_domain": types.StringType,
	"status":            types.StringType,
}

type CustomDomainList []CustomDomain

func (domains CustomDomainList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: customDomainAttrTypes,
		},
	}

	if domains == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(domains))
	for _, d := range domains {
		set.Elems = append(set.Elems, d.toTerraformObject())
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
		if d.Domain.Value == domain {
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
		if !oldDomains.contains(d) {
			diff.Create = append(diff.Create, d.toCreateRequest())
		}
	}

	return diff
}

type CustomDomain struct {
	Id               types.String `tfsdk:"id"`
	Domain           types.String `tfsdk:"domain"`
	ValidationDomain types.String `tfsdk:"validation_domain"`
	Status           types.String `tfsdk:"status"`
}

func (d CustomDomain) toTerraformObject() types.Object {
	return types.Object{
		AttrTypes: customDomainAttrTypes,
		Attrs: map[string]attr.Value{
			"id":                d.Id,
			"domain":            d.Domain,
			"validation_domain": d.ValidationDomain,
			"status":            d.Status,
		},
	}
}

func (d CustomDomain) toCreateRequest() client.CustomDomainCreateRequest {
	return client.CustomDomainCreateRequest{
		CustomDomainRequest: qovery.CustomDomainRequest{
			Domain: ToString(d.Domain),
		},
	}
}

func (d CustomDomain) toUpdateRequest(new CustomDomain) client.CustomDomainUpdateRequest {
	return client.CustomDomainUpdateRequest{
		Id: ToString(d.Id),
		CustomDomainRequest: qovery.CustomDomainRequest{
			Domain: ToString(new.Domain),
		},
	}
}

func (d CustomDomain) toDeleteRequest() client.CustomDomainDeleteRequest {
	return client.CustomDomainDeleteRequest{
		Id: ToString(d.Id),
	}
}

func fromCustomDomain(d *qovery.CustomDomain) CustomDomain {
	return CustomDomain{
		Id:               FromString(d.Id),
		Domain:           FromString(d.Domain),
		ValidationDomain: FromStringPointer(d.ValidationDomain),
		Status:           fromClientEnumPointer(d.Status),
	}
}

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
		Id:               v.Attrs["id"].(types.String),
		Domain:           v.Attrs["domain"].(types.String),
		ValidationDomain: v.Attrs["validation_domain"].(types.String),
		Status:           v.Attrs["status"].(types.String),
	}
}

func toCustomDomainList(vars types.Set) CustomDomainList {
	if vars.Null || vars.Unknown {
		return nil
	}

	customDomains := make([]CustomDomain, 0, len(vars.Elems))
	for _, elem := range vars.Elems {
		customDomains = append(customDomains, toCustomDomain(elem.(types.Object)))
	}

	return customDomains
}
