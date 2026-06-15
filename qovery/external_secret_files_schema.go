package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func externalSecretFilesSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description:         fmt.Sprintf("List of external secret files linked to this %s.", resourceType),
		MarkdownDescription: fmt.Sprintf("List of external secret files linked to this %s. External secret files reference upstream secrets (e.g. from AWS Secrets Manager) and are mounted as files at a given path inside the container.", resourceType),
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description:         "Id of the external secret file.",
					MarkdownDescription: "Id of the external secret file.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"key": schema.StringAttribute{
					Description:         "Name of the external secret file.",
					MarkdownDescription: "Name of the external secret file.",
					Required:            true,
				},
				"description": schema.StringAttribute{
					Description:         "Description of the external secret file.",
					MarkdownDescription: "Description of the external secret file.",
					Optional:            true,
				},
				"mount_path": schema.StringAttribute{
					Description:         "Absolute path where the secret file will be mounted inside the container.",
					MarkdownDescription: "Absolute path where the secret file will be mounted inside the container.",
					Required:            true,
				},
				"reference": schema.StringAttribute{
					Description:         "Reference to the upstream secret (e.g. the secret name or ARN in AWS Secrets Manager).",
					MarkdownDescription: "Reference to the upstream secret (e.g. the secret name or ARN in AWS Secrets Manager).",
					Required:            true,
				},
				"secret_manager_access_id": schema.StringAttribute{
					Description:         "Id of the secret manager access to use for this external secret file.",
					MarkdownDescription: "Id of the secret manager access to use for this external secret file.",
					Required:            true,
				},
			},
		},
	}
}
