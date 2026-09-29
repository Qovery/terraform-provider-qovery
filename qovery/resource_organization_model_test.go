//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func TestOrganizationStateFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName          string
		APIDescription    *string
		PriorDescription  types.String
		ExpectDescription types.String
	}{
		{TestName: "api_value_wins", APIDescription: new("set from the Console"), PriorDescription: types.StringValue("old"), ExpectDescription: types.StringValue("set from the Console")},
		{TestName: "a_description_cleared_from_the_console_shows_up", APIDescription: nil, PriorDescription: types.StringValue("old"), ExpectDescription: types.StringNull()},
		{TestName: "an_empty_api_description_keeps_a_null_prior", APIDescription: new(""), PriorDescription: types.StringNull(), ExpectDescription: types.StringNull()},
		{TestName: "import_records_the_remote_value", APIDescription: new("set from the Console"), PriorDescription: types.StringNull(), ExpectDescription: types.StringValue("set from the Console")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			orga := &organization.Organization{ID: uuid.New(), Name: "organization", Plan: organization.PlanEnterprise, Description: tc.APIDescription}

			got := organizationStateFromAPI(orga, Organization{Description: tc.PriorDescription})

			assert.Equal(t, orga.ID.String(), got.Id.ValueString())
			assert.Equal(t, "organization", got.Name.ValueString())
			assert.Equal(t, tc.ExpectDescription, got.Description)
		})
	}
}
