//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewLabelsGroupService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  labels_group.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.LabelsGroupRepository{},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewLabelsGroupService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestLabelsGroupService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := labels_group.UpsertServiceRequest{
		LabelsGroupUpsertRequest: labels_group.UpsertRequest{
			Name: gofakeit.Word(),
			Labels: []labels_group.LabelUpsertRequest{
				{Key: "key1", Value: "value1", PropagateToCloudProvider: true},
			},
		},
	}

	expectedResult := &labels_group.LabelsGroup{
		Id:   uuid.New(),
		Name: validRequest.LabelsGroupUpsertRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        labels_group.UpsertServiceRequest
		SetupMock      func(*mocks_test.LabelsGroupRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.LabelsGroupUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.LabelsGroupUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.LabelsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewLabelsGroupService(mockRepo)
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.OrganizationID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.Id, result.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLabelsGroupService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidGroupID := "invalid-uuid"
	emptyGroupID := ""

	expectedResult := &labels_group.LabelsGroup{
		Id:   uuid.MustParse(validGroupID),
		Name: gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		LabelsGroupID  string
		SetupMock      func(*mocks_test.LabelsGroupRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_labels_group_id",
			OrganizationID: validOrgID,
			LabelsGroupID:  emptyGroupID,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_invalid_labels_group_id",
			OrganizationID: validOrgID,
			LabelsGroupID:  invalidGroupID,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validGroupID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validGroupID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.LabelsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewLabelsGroupService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.OrganizationID, tc.LabelsGroupID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.Id, result.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLabelsGroupService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidGroupID := "invalid-uuid"

	validRequest := labels_group.UpsertServiceRequest{
		LabelsGroupUpsertRequest: labels_group.UpsertRequest{
			Name: gofakeit.Word(),
			Labels: []labels_group.LabelUpsertRequest{
				{Key: "key1", Value: "value1", PropagateToCloudProvider: true},
			},
		},
	}

	expectedResult := &labels_group.LabelsGroup{
		Id:   uuid.MustParse(validGroupID),
		Name: validRequest.LabelsGroupUpsertRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		LabelsGroupID  string
		Request        labels_group.UpsertServiceRequest
		SetupMock      func(*mocks_test.LabelsGroupRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			LabelsGroupID:  validGroupID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_invalid_labels_group_id",
			OrganizationID: validOrgID,
			LabelsGroupID:  invalidGroupID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validGroupID, validRequest.LabelsGroupUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validGroupID, validRequest.LabelsGroupUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.LabelsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewLabelsGroupService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.OrganizationID, tc.LabelsGroupID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.Id, result.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLabelsGroupService_Delete(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidGroupID := "invalid-uuid"
	emptyGroupID := ""

	testCases := []struct {
		TestName       string
		OrganizationID string
		LabelsGroupID  string
		SetupMock      func(*mocks_test.LabelsGroupRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_labels_group_id",
			OrganizationID: validOrgID,
			LabelsGroupID:  emptyGroupID,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to delete",
		},
		{
			TestName:       "error_invalid_labels_group_id",
			OrganizationID: validOrgID,
			LabelsGroupID:  invalidGroupID,
			SetupMock:      func(m *mocks_test.LabelsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to delete",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validGroupID).
					Return(errors.New("repository error"))
			},
			ExpectError: true,
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			LabelsGroupID:  validGroupID,
			SetupMock: func(m *mocks_test.LabelsGroupRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validGroupID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.LabelsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewLabelsGroupService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.OrganizationID, tc.LabelsGroupID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
