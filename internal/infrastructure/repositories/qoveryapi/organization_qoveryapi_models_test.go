package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func TestNewDomainOrganizationFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Organization  *qovery.Organization
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_credentials",
			Organization:  nil,
			ExpectedError: organization.ErrNilOrganization,
		},
		{
			TestName: "success",
			Organization: &qovery.Organization{
				Id:   gofakeit.UUID(),
				Name: gofakeit.Name(),
				Plan: qovery.PLANENUM_FREE,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			orga, err := newDomainOrganizationFromQovery(tc.Organization)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, orga)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, orga)
			assert.True(t, orga.IsValid())
			assert.Equal(t, tc.Organization.Id, orga.ID.String())
			assert.Equal(t, tc.Organization.Name, orga.Name)
			assert.Equal(t, string(tc.Organization.Plan), orga.Plan.String())
		})
	}
}

func TestNewQoveryOrganizationEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  organization.UpdateRequest
	}{
		{
			TestName: "success_without_description",
			Request: organization.UpdateRequest{
				Name: gofakeit.Name(),
			},
		},
		{
			TestName: "success_with_description",
			Request: organization.UpdateRequest{
				Name:        gofakeit.Name(),
				Description: new(gofakeit.Word()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryOrganizationEditRequestFromDomain(tc.Request)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.Description, req.Description)
		})
	}
}
