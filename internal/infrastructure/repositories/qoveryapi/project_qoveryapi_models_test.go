package qoveryapi

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
)

func TestNewDomainProjectFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Project       *qovery.Project
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_project",
			Project:       nil,
			ExpectedError: project.ErrNilProject,
		},
		{
			TestName: "success",
			Project: &qovery.Project{
				Id: gofakeit.UUID(),
				Organization: &qovery.ReferenceObject{
					Id: gofakeit.UUID(),
				},
				Name:        gofakeit.Name(),
				Description: pointer.ToString(gofakeit.Name()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := newDomainProjectFromQovery(tc.Project)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, proj)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, proj)
			assert.True(t, proj.IsValid())
			assert.Equal(t, tc.Project.Id, proj.ID.String())
			assert.Equal(t, tc.Project.Organization.Id, proj.OrganizationID.String())
			assert.Equal(t, tc.Project.Name, proj.Name)
			assert.Equal(t, tc.Project.Description, proj.Description)
		})
	}
}

func TestNewQoveryProjectEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  project.UpsertRepositoryRequest
	}{
		{
			TestName: "success_without_description",
			Request: project.UpsertRepositoryRequest{
				Name: gofakeit.Name(),
			},
		},
		{
			TestName: "success_with_description",
			Request: project.UpsertRepositoryRequest{
				Name:        gofakeit.Name(),
				Description: pointer.ToString(gofakeit.Word()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryProjectRequestFromDomain(tc.Request)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.Description, req.Description)
		})
	}
}
