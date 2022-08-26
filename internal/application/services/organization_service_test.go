//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type OrganizationServiceTestSuite struct {
	suite.Suite

	repository *mocks_test.OrganizationRepository
	service    organization.Service
}

func (ts *OrganizationServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize repository
	organizationRepository := mocks_test.NewOrganizationRepository(t)

	// Initialize service
	organizationService, err := services.NewOrganizationService(organizationRepository)
	require.NoError(t, err)
	require.NotNil(t, organizationService)

	ts.repository = organizationRepository
	ts.service = organizationService
}

func (ts *OrganizationServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	organizationService, err := services.NewOrganizationService(nil)
	assert.Nil(t, organizationService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *OrganizationServiceTestSuite) TestNew_Success() {
	t := ts.T()

	organizationService, err := services.NewOrganizationService(mocks_test.NewOrganizationRepository(t))
	assert.Nil(t, err)
	assert.NotNil(t, organizationService)
}

func (ts *OrganizationServiceTestSuite) TestGet_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			orga, err := ts.service.Get(context.Background(), tc.OrganizationID)
			assert.Nil(t, orga)
			assert.ErrorContains(t, err, organization.ErrFailedToGetOrganization.Error())
			assert.ErrorContains(t, err, organization.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *OrganizationServiceTestSuite) TestGet_FailWithMissingOrganization() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID).
		Return(nil, errors.New(""))

	orga, err := ts.service.Get(context.Background(), fakeID)
	assert.Nil(t, orga)
	assert.ErrorContains(t, err, organization.ErrFailedToGetOrganization.Error())
}

func (ts *OrganizationServiceTestSuite) TestGet_Success() {
	t := ts.T()

	expectedOrga := assertCreateOrganization(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedOrga.ID.String()).
		Return(expectedOrga, nil)

	orga, err := ts.service.Get(context.Background(), expectedOrga.ID.String())
	assert.Nil(t, err)
	assertEqualOrganization(t, expectedOrga, orga)
}

func (ts *OrganizationServiceTestSuite) TestUpdate_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			orga, err := ts.service.Update(context.Background(), tc.OrganizationID, assertNewOrganizationUpdateRequest(t))
			assert.Nil(t, orga)
			assert.ErrorContains(t, err, organization.ErrFailedToUpdateOrganization.Error())
			assert.ErrorContains(t, err, organization.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *OrganizationServiceTestSuite) TestUpdate_FailOrganizationNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewOrganizationUpdateRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, updateRequest).
		Return(nil, errors.New(""))

	orga, err := ts.service.Update(context.Background(), fakeID, updateRequest)
	assert.Nil(t, orga)
	assert.ErrorContains(t, err, organization.ErrFailedToUpdateOrganization.Error())
}

func (ts *OrganizationServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	expectedOrga := assertCreateOrganization(t)
	updateRequest := organization.UpdateRequest{}

	orga, err := ts.service.Update(context.Background(), expectedOrga.ID.String(), updateRequest)
	assert.Nil(t, orga)
	assert.ErrorContains(t, err, organization.ErrFailedToUpdateOrganization.Error())
	assert.ErrorContains(t, err, organization.ErrInvalidUpdateRequest.Error())
}

func (ts *OrganizationServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	newOrga := assertCreateOrganization(t)
	updateRequest := assertNewOrganizationUpdateRequest(t)
	updatedOrga := assertApplyOrganizationUpdateRequest(t, newOrga, updateRequest)

	ts.repository.EXPECT().
		Update(mock.Anything, updatedOrga.ID.String(), updateRequest).
		Return(updatedOrga, nil)

	orga, err := ts.service.Update(context.Background(), updatedOrga.ID.String(), updateRequest)
	assert.Nil(t, err)
	assertEqualOrganization(t, updatedOrga, orga)
}

func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}

func assertNewOrganizationUpdateRequest(t *testing.T) organization.UpdateRequest {
	req := organization.UpdateRequest{
		Name:        gofakeit.Name(),
		Description: pointer.ToString(gofakeit.Word()),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertCreateOrganization(t *testing.T) *organization.Organization {
	orga, err := organization.NewOrganization(organization.NewOrganizationParams{
		OrganizationID: gofakeit.UUID(),
		Name:           gofakeit.Name(),
		Plan:           organization.PlanFree.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, orga)
	require.NoError(t, orga.Validate())

	return orga
}

func assertEqualOrganization(t *testing.T, expected *organization.Organization, actual *organization.Organization) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.Plan, actual.Plan)
}

func assertApplyOrganizationUpdateRequest(t *testing.T, orga *organization.Organization, update organization.UpdateRequest) *organization.Organization {
	updatedOrga, err := organization.NewOrganization(organization.NewOrganizationParams{
		OrganizationID: orga.ID.String(),
		Plan:           orga.Plan.String(),
		Name:           update.Name,
		Description:    update.Description,
	})
	require.NoError(t, err)
	require.NotNil(t, updatedOrga)
	require.NoError(t, updatedOrga.Validate())

	return updatedOrga
}
