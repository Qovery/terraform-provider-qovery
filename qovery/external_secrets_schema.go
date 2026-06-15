package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func externalSecretsSchemaAttribute(resourceType string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description:         fmt.Sprintf("List of external secrets linked to this %s.", resourceType),
		MarkdownDescription: fmt.Sprintf("List of external secrets linked to this %s. External secrets reference upstream secrets (e.g. from AWS Secrets Manager) via a secret manager access configuration.", resourceType),
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description:         "Id of the external secret.",
					MarkdownDescription: "Id of the external secret.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"key": schema.StringAttribute{
					Description:         "Name of the external secret.",
					MarkdownDescription: "Name of the external secret.",
					Required:            true,
				},
				"description": schema.StringAttribute{
					Description:         "Description of the external secret.",
					MarkdownDescription: "Description of the external secret.",
					Optional:            true,
				},
				"reference": schema.StringAttribute{
					Description:         "Reference to the upstream secret (e.g. the secret name or ARN in AWS Secrets Manager).",
					MarkdownDescription: "Reference to the upstream secret (e.g. the secret name or ARN in AWS Secrets Manager).",
					Required:            true,
				},
				"secret_manager_access_id": schema.StringAttribute{
					Description:         "Id of the secret manager access to use for this external secret.",
					MarkdownDescription: "Id of the secret manager access to use for this external secret.",
					Required:            true,
				},
			},
		},
	}
}
