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
		Current  qovery.Organization
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
		{
			TestName: "success_copies_unmanaged_fields_from_current",
			Request: organization.UpdateRequest{
				Name:        gofakeit.Name(),
				Description: new(gofakeit.Word()),
			},
			Current: qovery.Organization{
				Name:        gofakeit.Name(),
				Description: *qovery.NewNullableString(new(gofakeit.Word())),
				WebsiteUrl:  *qovery.NewNullableString(new(gofakeit.URL())),
				Repository:  *qovery.NewNullableString(new(gofakeit.URL())),
				LogoUrl:     *qovery.NewNullableString(new(gofakeit.URL())),
				IconUrl:     *qovery.NewNullableString(new(gofakeit.URL())),
				AdminEmails: []string{gofakeit.Email(), gofakeit.Email()},
			},
		},
		{
			TestName: "success_copies_null_unmanaged_fields_from_current",
			Request: organization.UpdateRequest{
				Name: gofakeit.Name(),
			},
			Current: qovery.Organization{
				WebsiteUrl:  *qovery.NewNullableString(nil),
				Repository:  *qovery.NewNullableString(nil),
				LogoUrl:     *qovery.NewNullableString(nil),
				IconUrl:     *qovery.NewNullableString(nil),
				AdminEmails: []string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryOrganizationEditRequestFromDomain(tc.Request, tc.Current)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.Description, req.Description)
			assert.Equal(t, tc.Current.WebsiteUrl, req.WebsiteUrl)
			assert.Equal(t, tc.Current.Repository, req.Repository)
			assert.Equal(t, tc.Current.LogoUrl, req.LogoUrl)
			assert.Equal(t, tc.Current.IconUrl, req.IconUrl)
			assert.Equal(t, tc.Current.AdminEmails, req.AdminEmails)
		})
	}
}
