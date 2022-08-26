package project_test

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
)

func TestNewProject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        project.NewProjectParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_project_id",
			Params: project.NewProjectParams{
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
				Description:    pointer.ToString(gofakeit.Name()),
			},
			ExpectedError: project.ErrInvalidProjectIDParam,
		},
		{
			TestName: "fail_with_invalid_organization_id",
			Params: project.NewProjectParams{
				ProjectID:   gofakeit.UUID(),
				Name:        gofakeit.Name(),
				Description: pointer.ToString(gofakeit.Name()),
			},
			ExpectedError: project.ErrInvalidProjectOrganizationIDParam,
		},
		{
			TestName: "fail_with_invalid_name",
			Params: project.NewProjectParams{
				ProjectID:      gofakeit.UUID(),
				OrganizationID: gofakeit.UUID(),
				Description:    pointer.ToString(gofakeit.Name()),
			},
			ExpectedError: project.ErrInvalidProjectNameParam,
		},
		{
			TestName: "success_without_description",
			Params: project.NewProjectParams{
				ProjectID:      gofakeit.UUID(),
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
			},
		},
		{
			TestName: "success_with_description",
			Params: project.NewProjectParams{
				ProjectID:      gofakeit.UUID(),
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
				Description:    pointer.ToString(gofakeit.Name()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := project.NewProject(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, proj)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, proj)
			assert.True(t, proj.IsValid())
			assert.Equal(t, tc.Params.ProjectID, proj.ID.String())
			assert.Equal(t, tc.Params.OrganizationID, proj.OrganizationID.String())
			assert.Equal(t, tc.Params.Name, proj.Name)
			assert.Equal(t, tc.Params.Description, proj.Description)
			assert.Len(t, tc.Params.EnvironmentVariables, len(proj.EnvironmentVariables))
		})
	}
}
