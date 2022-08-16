package organization_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func TestNewOrganization(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        organization.NewOrganizationParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_organization_id",
			Params: organization.NewOrganizationParams{
				Name: gofakeit.Name(),
				Plan: organization.PlanFree.String(),
			},
			ExpectedError: organization.ErrInvalidOrganizationIDParam,
		},
		{
			TestName: "fail_with_invalid_name",
			Params: organization.NewOrganizationParams{
				OrganizationID: gofakeit.UUID(),
				Plan:           organization.PlanFree.String(),
			},
			ExpectedError: organization.ErrInvalidNameParam,
		},
		{
			TestName: "fail_with_invalid_plan",
			Params: organization.NewOrganizationParams{
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
			},
			ExpectedError: organization.ErrInvalidPlanParam,
		},
		{
			TestName: "success",
			Params: organization.NewOrganizationParams{
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
				Plan:           organization.PlanFree.String(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			orga, err := organization.NewOrganization(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, orga)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, orga)
			assert.Equal(t, tc.Params.OrganizationID, orga.ID.String())
			assert.Equal(t, tc.Params.Name, orga.Name)
			assert.Equal(t, tc.Params.Plan, orga.Plan.String())
		})
	}
}
