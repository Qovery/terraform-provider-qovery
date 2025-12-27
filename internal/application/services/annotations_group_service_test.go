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

	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewAnnotationsGroupService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  annotations_group.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.AnnotationsGroupRepository{},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewAnnotationsGroupService(tc.Repository)
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

func TestAnnotationsGroupService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := annotations_group.UpsertServiceRequest{
		AnnotationsGroupUpsertRequest: annotations_group.UpsertRequest{
			Name: gofakeit.Word(),
			Annotations: []annotations_group.AnnotationUpsertRequest{
				{Key: "key1", Value: "value1"},
			},
			Scopes: []string{"PODS"},
		},
	}

	expectedResult := &annotations_group.AnnotationsGroup{
		Id:   uuid.New(),
		Name: validRequest.AnnotationsGroupUpsertRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        annotations_group.UpsertServiceRequest
		SetupMock      func(*mocks_test.AnnotationsGroupRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.AnnotationsGroupUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.AnnotationsGroupUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.AnnotationsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewAnnotationsGroupService(mockRepo)
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

func TestAnnotationsGroupService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidGroupID := "invalid-uuid"
	emptyGroupID := ""

	expectedResult := &annotations_group.AnnotationsGroup{
		Id:   uuid.MustParse(validGroupID),
		Name: gofakeit.Word(),
	}

	testCases := []struct {
		TestName         string
		OrganizationID   string
		AnnotationsGrpID string
		SetupMock        func(*mocks_test.AnnotationsGroupRepository)
		ExpectError      bool
		ErrorContains    string
	}{
		{
			TestName:         "error_empty_annotations_group_id",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: emptyGroupID,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "invalid",
		},
		{
			TestName:         "error_invalid_annotations_group_id",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: invalidGroupID,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "invalid",
		},
		{
			TestName:         "error_repository_failure",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validGroupID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get",
		},
		{
			TestName:         "success",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
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

			mockRepo := &mocks_test.AnnotationsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewAnnotationsGroupService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.OrganizationID, tc.AnnotationsGrpID)

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

func TestAnnotationsGroupService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidGroupID := "invalid-uuid"

	validRequest := annotations_group.UpsertServiceRequest{
		AnnotationsGroupUpsertRequest: annotations_group.UpsertRequest{
			Name: gofakeit.Word(),
			Annotations: []annotations_group.AnnotationUpsertRequest{
				{Key: "key1", Value: "value1"},
			},
			Scopes: []string{"PODS"},
		},
	}

	expectedResult := &annotations_group.AnnotationsGroup{
		Id:   uuid.MustParse(validGroupID),
		Name: validRequest.AnnotationsGroupUpsertRequest.Name,
	}

	testCases := []struct {
		TestName         string
		OrganizationID   string
		AnnotationsGrpID string
		Request          annotations_group.UpsertServiceRequest
		SetupMock        func(*mocks_test.AnnotationsGroupRepository)
		ExpectError      bool
		ErrorContains    string
	}{
		{
			TestName:         "error_invalid_organization_id",
			OrganizationID:   invalidOrgID,
			AnnotationsGrpID: validGroupID,
			Request:          validRequest,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "invalid",
		},
		{
			TestName:         "error_invalid_annotations_group_id",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: invalidGroupID,
			Request:          validRequest,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "invalid",
		},
		{
			TestName:         "error_repository_failure",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			Request:          validRequest,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validGroupID, validRequest.AnnotationsGroupUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update",
		},
		{
			TestName:         "success",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			Request:          validRequest,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validGroupID, validRequest.AnnotationsGroupUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.AnnotationsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewAnnotationsGroupService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.OrganizationID, tc.AnnotationsGrpID, tc.Request)

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

func TestAnnotationsGroupService_Delete(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validGroupID := gofakeit.UUID()
	invalidGroupID := "invalid-uuid"
	emptyGroupID := ""

	testCases := []struct {
		TestName         string
		OrganizationID   string
		AnnotationsGrpID string
		SetupMock        func(*mocks_test.AnnotationsGroupRepository)
		ExpectError      bool
		ErrorContains    string
	}{
		{
			TestName:         "error_empty_annotations_group_id",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: emptyGroupID,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "failed to delete",
		},
		{
			TestName:         "error_invalid_annotations_group_id",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: invalidGroupID,
			SetupMock:        func(m *mocks_test.AnnotationsGroupRepository) {},
			ExpectError:      true,
			ErrorContains:    "failed to delete",
		},
		{
			TestName:         "error_repository_failure",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validGroupID).
					Return(errors.New("repository error"))
			},
			ExpectError: true,
		},
		{
			TestName:         "success",
			OrganizationID:   validOrgID,
			AnnotationsGrpID: validGroupID,
			SetupMock: func(m *mocks_test.AnnotationsGroupRepository) {
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

			mockRepo := &mocks_test.AnnotationsGroupRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewAnnotationsGroupService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.OrganizationID, tc.AnnotationsGrpID)

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
