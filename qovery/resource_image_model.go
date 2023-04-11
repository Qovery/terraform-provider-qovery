package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
)

type Image struct {
	RegistryID types.String `tfsdk:"registry_id"`
	Name       types.String `tfsdk:"name"`
	Tag        types.String `tfsdk:"tag"`
}

func (i Image) toUpsertRequest() *image.Image {
	return &image.Image{
		RegistryID: ToString(i.RegistryID),
		Name:       ToString(i.Name),
		Tag:        ToString(i.Tag),
	}
}
