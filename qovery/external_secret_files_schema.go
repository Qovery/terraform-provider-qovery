package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func externalSecretFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	d := variableListDescriptions("external_secret_files", resourceType)
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
				"mount_path": schema.StringAttribute{
					MarkdownDescription: d.MountPath,
					Required:            true,
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
