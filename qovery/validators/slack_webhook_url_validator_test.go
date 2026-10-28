//go:build unit && !integration
// +build unit,!integration

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSlackWebhookURLValidator_ValidateString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		ConfigValue types.String
		ExpectError bool
	}{
		{TestName: "success_slack_webhook_url", ConfigValue: types.StringValue("https://hooks.slack.com/services/T000/B000/secret")},
		// The mask the API returns keeps the current webhook.
		{TestName: "success_mask", ConfigValue: types.StringValue("https://hooks.slack.com/services/XXXXXXXXXXX")},
		{TestName: "success_uppercase_scheme_and_host", ConfigValue: types.StringValue("HTTPS://HOOKS.SLACK.COM/services/T000/B000/secret")},
		{TestName: "success_explicit_port", ConfigValue: types.StringValue("https://hooks.slack.com:443/services/T000/B000/secret")},
		{TestName: "success_user_info", ConfigValue: types.StringValue("https://user@hooks.slack.com/services/T000/B000/secret")},
		{TestName: "success_host_without_path", ConfigValue: types.StringValue("https://hooks.slack.com")},
		{TestName: "success_null", ConfigValue: types.StringNull()},
		{TestName: "success_unknown", ConfigValue: types.StringUnknown()},
		{TestName: "error_empty", ConfigValue: types.StringValue(""), ExpectError: true},
		{TestName: "error_http_scheme", ConfigValue: types.StringValue("http://hooks.slack.com/services/T000/B000/secret"), ExpectError: true},
		{TestName: "error_host_suffixed_by_another_domain", ConfigValue: types.StringValue("https://hooks.slack.com.evil.com/services/secret"), ExpectError: true},
		{TestName: "error_slack_host_in_the_path", ConfigValue: types.StringValue("https://evil.com/hooks.slack.com/secret"), ExpectError: true},
		{TestName: "error_other_slack_host", ConfigValue: types.StringValue("https://slack.com/services/T000/B000/secret"), ExpectError: true},
		{TestName: "error_without_scheme", ConfigValue: types.StringValue("hooks.slack.com/services/secret"), ExpectError: true},
		{TestName: "error_unparsable", ConfigValue: types.StringValue("https://hooks.slack.com/%zz-secret"), ExpectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("using_wide_credentials").AtName("slack_webhook_url"),
				ConfigValue: tc.ConfigValue,
			}
			resp := &validator.StringResponse{}
			NewSlackWebhookURLValidator().ValidateString(context.Background(), req, resp)

			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError())
			// The URL is a secret: no diagnostic shows it.
			if value := tc.ConfigValue.ValueString(); value != "" {
				for _, d := range resp.Diagnostics {
					assert.NotContains(t, d.Summary(), value)
					assert.NotContains(t, d.Detail(), value)
				}
			}
		})
	}
}
