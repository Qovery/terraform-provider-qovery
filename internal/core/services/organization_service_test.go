//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/core/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func TestNewOrganizationScalewayService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    organization.Repository
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_repository",
			Repository:    nil,
			ExpectedError: services.ErrInvalidRepository,
		},
		{
			TestName:   "success_with_repository",
			Repository: mocks_test.NewOrganizationRepository(t),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			svc, err := services.NewOrganizationService(tc.Repository)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, svc)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, svc)
		})
	}
}

func TestOrganizationService_Get(t *testing.T) {
	t.Parallel()

	// Initialize service
	newOrga, err := organization.NewOrganization(organization.NewOrganizationParams{
		OrganizationID: gofakeit.UUID(),
		Name:           gofakeit.Name(),
		Plan:           organization.PlanFree.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, newOrga)

	orgaRepo := mocks_test.NewOrganizationRepository(t)
	orgaRepo.EXPECT().
		Get(mock.Anything, newOrga.ID.String()).
		Return(newOrga, nil)
	orgaRepo.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(nil, organization.ErrFailedToGetOrganization)

	organizationService, err := services.NewOrganizationService(orgaRepo)
	require.NoError(t, err)
	require.NotNil(t, organizationService)

	testCases := []struct {
		TestName       string
		OrganizationID string
		Expected       *organization.Organization
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "",
			ExpectedError:  organization.ErrInvalidOrganizationIDParam,
		},
		{
			TestName:       "fail_not_found",
			OrganizationID: gofakeit.UUID(),
			ExpectedError:  organization.ErrFailedToGetOrganization,
		},
		{
			TestName:       "success",
			OrganizationID: newOrga.ID.String(),
			Expected:       newOrga,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			orga, err := organizationService.Get(context.Background(), tc.OrganizationID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, orga)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, orga)
			assert.Equal(t, tc.Expected.ID, orga.ID)
			assert.Equal(t, tc.Expected.Name, orga.Name)
		})
	}
}

//func TestOrganizationService_Update(t *testing.T) {
//	t.Parallel()
//
//	// Initialize service
//	awsOrganizationService, err := services.NewOrganizationService(inmem.NewOrganizationInmem())
//	require.NoError(t, err)
//	require.NotNil(t, awsOrganizationService)
//
//	newCreds, err := awsOrganizationService.Create(context.Background(), uuid.NewString(), organization.UpsertAwsRequest{
//		Name:            gofakeit.Name(),
//		SecretAccessKey: gofakeit.Word(),
//		AccessKeyID:     gofakeit.Word(),
//	})
//	require.NoError(t, err)
//	require.NotNil(t, newCreds)
//
//	updateRequest := organization.UpsertAwsRequest{
//		Name:            gofakeit.Name(),
//		SecretAccessKey: gofakeit.Word(),
//		AccessKeyID:     gofakeit.Word(),
//	}
//
//	testCases := []struct {
//		TestName       string
//		OrganizationID string
//		OrganizationID string
//		Request        organization.UpsertAwsRequest
//		Expected       *organization.Organization
//		ExpectedError  error
//	}{
//		{
//			TestName:       "fail_with_invalid_organization_id",
//			OrganizationID: "",
//			OrganizationID: newCreds.ID.String(),
//			ExpectedError:  organization.ErrInvalidOrganizationIDParam,
//		},
//		{
//			TestName:       "fail_with_invalid_organization_id",
//			OrganizationID: newCreds.OrganizationID.String(),
//			OrganizationID: "",
//			ExpectedError:  organization.ErrInvalidOrganizationIDParam,
//		},
//		{
//			TestName:       "fail_with_invalid_upsert_aws_request",
//			OrganizationID: newCreds.OrganizationID.String(),
//			OrganizationID: newCreds.ID.String(),
//			ExpectedError:  organization.ErrInvalidUpsertAwsRequest,
//		},
//		{
//			TestName:       "success",
//			OrganizationID: newCreds.OrganizationID.String(),
//			OrganizationID: newCreds.ID.String(),
//			Request:        updateRequest,
//			Expected: &organization.Organization{
//				ID:             newCreds.ID,
//				OrganizationID: newCreds.OrganizationID,
//				Name:           updateRequest.Name,
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		t.Run(tc.TestName, func(t *testing.T) {
//			orga, err := awsOrganizationService.Update(context.Background(), tc.OrganizationID, tc.OrganizationID, tc.Request)
//			if tc.ExpectedError != nil {
//				assert.ErrorContains(t, err, tc.ExpectedError.Error())
//				assert.ErrorContains(t, err, organization.ErrFailedToUpdateAwsOrganization.Error())
//				assert.Nil(t, orga)
//				return
//			}
//
//			assert.NoError(t, err)
//			assert.NotNil(t, orga)
//			assert.Equal(t, tc.Expected.ID, orga.ID)
//			assert.Equal(t, tc.Expected.OrganizationID, orga.OrganizationID)
//			assert.Equal(t, tc.Expected.Name, orga.Name)
//		})
//	}
//}
