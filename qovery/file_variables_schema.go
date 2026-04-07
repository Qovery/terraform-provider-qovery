package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// environmentVariableFilesSchemaAttribute returns the schema for environment_variable_files,
// parameterized by resource type name for the description.
func environmentVariableFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: fmt.Sprintf("List of environment variable files linked to this %s.", resourceType),
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "Id of the environment variable file.",
					Computed:    true,
				},
				"key": schema.StringAttribute{
					Description: "Key of the environment variable file.",
					Required:    true,
				},
				"value": schema.StringAttribute{
					Description: "Value of the environment variable file.",
					Required:    true,
				},
				"mount_path": schema.StringAttribute{
					Description: "Mount path of the environment variable file.",
					Required:    true,
				},
				"description": schema.StringAttribute{
					Description: "Description of the environment variable file.",
					Optional:    true,
				},
			},
		},
	}
}

// secretFilesSchemaAttribute returns the schema for secret_files,
// parameterized by resource type name for the description.
func secretFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: fmt.Sprintf("List of secret files linked to this %s.", resourceType),
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "Id of the secret file.",
					Computed:    true,
				},
				"key": schema.StringAttribute{
					Description: "Key of the secret file.",
					Required:    true,
				},
				"value": schema.StringAttribute{
					Description: "Value of the secret file.",
					Required:    true,
					Sensitive:   true,
				},
				"mount_path": schema.StringAttribute{
					Description: "Mount path of the secret file.",
					Required:    true,
				},
				"description": schema.StringAttribute{
					Description: "Description of the secret file.",
					Optional:    true,
				},
			},
		},
	}
}
