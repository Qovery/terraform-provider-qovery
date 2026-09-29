package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// environmentVariableFilesSchemaAttribute returns the schema for environment_variable_files,
// parameterized by resource type name for the description.
func environmentVariableFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	d := variableListDescriptions("environment_variable_files", resourceType)
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
				"value": schema.StringAttribute{
					MarkdownDescription: d.Value,
					Required:            true,
				},
				"mount_path": schema.StringAttribute{
					MarkdownDescription: d.MountPath,
					Required:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: d.Description,
					Optional:            true,
				},
			},
		},
	}
}

// secretFilesSchemaAttribute returns the schema for secret_files,
// parameterized by resource type name for the description.
func secretFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	d := variableListDescriptions("secret_files", resourceType)
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
				"value": schema.StringAttribute{
					MarkdownDescription: d.Value,
					Required:            true,
					Sensitive:           true,
				},
				"mount_path": schema.StringAttribute{
					MarkdownDescription: d.MountPath,
					Required:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: d.Description,
					Optional:            true,
				},
			},
		},
	}
}
