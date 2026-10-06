//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// TestNewQoveryTerraformRequestFromDomain_AlwaysSentFields guards that description,
// auto_deploy_config, backend and engine stay in every request, an unset description
// as "", now that the client makes them optional.
func TestNewQoveryTerraformRequestFromDomain_AlwaysSentFields(t *testing.T) {
	t.Parallel()

	description := "my terraform service"
	testCases := []struct {
		TestName            string
		Description         *string
		TerraformAction     terraformservice.TerraformAction
		ExpectedDescription string
		ExpectedAction      terraformservice.TerraformAction
	}{
		{
			TestName:            "without_description",
			ExpectedDescription: "",
			ExpectedAction:      terraformservice.TerraformActionDefault,
		},
		{
			TestName:            "with_description",
			Description:         &description,
			TerraformAction:     terraformservice.TerraformActionPlan,
			ExpectedDescription: description,
			ExpectedAction:      terraformservice.TerraformActionPlan,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			request := validTerraformUpsertRequest()
			request.Description = tc.Description
			request.TerraformAction = tc.TerraformAction

			req, err := newQoveryTerraformRequestFromDomain(request)
			require.NoError(t, err)
			body, err := json.Marshal(req)
			require.NoError(t, err)

			var fields map[string]any
			require.NoError(t, json.Unmarshal(body, &fields))
			assert.Equal(t, tc.ExpectedDescription, fields["description"])
			assert.Equal(t, map[string]any{"auto_deploy": false, "terraform_action": string(tc.ExpectedAction)}, fields["auto_deploy_config"])
			assert.Equal(t, map[string]any{"kubernetes": map[string]any{}}, fields["backend"])
			assert.Equal(t, string(terraformservice.EngineTerraform), fields["engine"])
			assert.NotContains(t, fields, "auto_deploy")
			assert.NotContains(t, fields, "auto_preview")
		})
	}
}
