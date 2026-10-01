package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func externalSecretsSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	d := variableListDescriptions("external_secrets", resourceType)
	return schema.SetNestedAttribute{
		MarkdownDescription: d.List,
		Optional:            true,
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
				"description": schema.StringAttribute{
					MarkdownDescription: d.Description,
					Optional:            true,
				},
				"reference": schema.StringAttribute{
					MarkdownDescription: d.Reference,
					Required:            true,
				},
				"secret_manager_access_id": schema.StringAttribute{
					MarkdownDescription: d.SecretManagerAccessID,
					Required:            true,
				},
			},
		},
	}
}
