package validators

import (
	"context"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	slackWebhookURLScheme = "https"
	slackWebhookURLHost   = "hooks.slack.com"
)

var _ validator.String = SlackWebhookURLValidator{}

// SlackWebhookURLValidator validates that a string is a Slack webhook URL: its scheme is https and
// its host hooks.slack.com, both case-insensitive. It runs the check q-core runs at apply
// (ClusterWidePermissionsDomain.validateSlackWebhookUrl), so plan rejects only what apply rejects.
// The URL is a secret: unlike the framework validators, the error never shows the value.
// Null and unknown values pass.
type SlackWebhookURLValidator struct{}

// NewSlackWebhookURLValidator returns a validator for a sensitive Slack webhook URL attribute.
func NewSlackWebhookURLValidator() SlackWebhookURLValidator {
	return SlackWebhookURLValidator{}
}

// Description returns a plain text description of the validator's behavior.
func (v SlackWebhookURLValidator) Description(_ context.Context) string {
	return "value must target `" + slackWebhookURLScheme + "://" + slackWebhookURLHost + "`"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v SlackWebhookURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString checks that a string attribute is a Slack webhook URL.
func (v SlackWebhookURLValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	parsed, err := url.Parse(req.ConfigValue.ValueString())
	if err == nil && strings.EqualFold(parsed.Scheme, slackWebhookURLScheme) && strings.EqualFold(parsed.Hostname(), slackWebhookURLHost) {
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Slack webhook URL",
		req.Path.String()+" must target `"+slackWebhookURLScheme+"://"+slackWebhookURLHost+"`. The value is not shown, as it is sensitive.",
	)
}
