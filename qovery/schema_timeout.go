package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

const timeoutDocumentationURL = "https://www.terraform.io/language/resources/syntax#operation-timeouts"

type Timeout struct {
	Create types.String `tfsdk:"create"`
	Update types.String `tfsdk:"update"`
	Delete types.String `tfsdk:"delete"`
}

type TimeoutParams struct {
	ResourceName  string
	CreateDefault string
	UpdateDefault string
	DeleteDefault string
}

func NewTimeoutSchemaAttribute(params TimeoutParams) tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: fmt.Sprintf(
			"`%s` provides the following [Timeouts](%s) configuration options.",
			params.ResourceName,
			timeoutDocumentationURL,
		),
		Optional: true,
		Computed: true,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"create": {
				Description: descriptions.NewStringDefaultDescription(
					fmt.Sprintf("Used for %s creation.", params.ResourceName),
					params.CreateDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(params.CreateDefault),
				},
			},
			"update": {
				Description: descriptions.NewStringDefaultDescription(
					fmt.Sprintf("Used for %s modifications.", params.ResourceName),
					params.UpdateDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(params.UpdateDefault),
				},
			},
			"delete": {
				Description: descriptions.NewStringDefaultDescription(
					fmt.Sprintf("Used for %s destructions.", params.ResourceName),
					params.DeleteDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(params.DeleteDefault),
				},
			},
		}),
	}
}
