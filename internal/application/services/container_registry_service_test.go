//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	mock_repository "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type RegistryServiceTestSuite struct {
	suite.Suite

	repository *mock_repository.RegistryRepository
	service    registry.Service
}

func (ts *RegistryServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize registry repository
	registryRepository := mock_repository.NewRegistryRepository(t)

	// Initialize registry service
	registryService, err := services.NewContainerRegistryService(registryRepository)
	require.NoError(t, err)
	require.NotNil(t, registryService)

	ts.repository = registryRepository
	ts.service = registryService
}

func (ts *RegistryServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	regService, err := services.NewContainerRegistryService(nil)
	assert.Nil(t, regService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *RegistryServiceTestSuite) TestNew_Success() {
	t := ts.T()

	regService, err := services.NewContainerRegistryService(ts.repository)
	assert.Nil(t, err)
	assert.NotNil(t, regService)
}

func (ts *RegistryServiceTestSuite) TestCreate_FailWithInvalidOrganizationID() {
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
			reg, err := ts.service.Create(context.Background(), tc.OrganizationID, assertNewRegistryUpsertRequest(t))
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToCreateRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest registry.UpsertRequest
	}{
		{
			TestName: "invalid_name",
			CreateRequest: registry.UpsertRequest{
				Kind: registry.KindDockerHub.String(),
				URL:  gofakeit.URL(),
			},
		},
		{
			TestName: "invalid_kind",
			CreateRequest: registry.UpsertRequest{
				Name: gofakeit.Name(),
				URL:  gofakeit.URL(),
			},
		},
		{
			TestName: "invalid_url",
			CreateRequest: registry.UpsertRequest{
				Name: gofakeit.Name(),
				Kind: registry.KindDockerHub.String(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			reg, err := ts.service.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToCreateRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestCreate_FailedToCreateRegistry() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	createRequest := assertNewRegistryUpsertRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, createRequest).
		Return(nil, registry.ErrInvalidRegistry)

	reg, err := ts.service.Create(context.Background(), organizationID, createRequest)
	assert.Nil(t, reg)
	assert.ErrorContains(t, err, registry.ErrFailedToCreateRegistry.Error())
	assert.ErrorContains(t, err, registry.ErrInvalidRegistry.Error())
}

func (ts *RegistryServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	expectedRegistry := assertCreateRegistry(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedRegistry.OrganizationID.String(), mock.Anything).
		Return(expectedRegistry, nil)

	reg, err := ts.service.Create(context.Background(), expectedRegistry.OrganizationID.String(), assertNewRegistryUpsertRequest(t))
	assert.Nil(t, err)
	assertEqualRegistry(t, expectedRegistry, reg)
}

func (ts *RegistryServiceTestSuite) TestGet_FailWithInvalidOrganizationID() {
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
			reg, err := ts.service.Get(context.Background(), tc.OrganizationID, gofakeit.UUID())
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToGetRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestGet_FailWithInvalidRegistryID() {
	t := ts.T()

	testCases := []struct {
		TestName   string
		RegistryID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:   "invalid_uuid",
			RegistryID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			reg, err := ts.service.Get(context.Background(), gofakeit.UUID(), tc.RegistryID)
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToGetRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidRegistryIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestGet_FailRegistryNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID, fakeID).
		Return(nil, registry.ErrInvalidRegistry)

	reg, err := ts.service.Get(context.Background(), fakeID, fakeID)
	assert.Nil(t, reg)
	assert.ErrorContains(t, err, registry.ErrFailedToGetRegistry.Error())
	assert.ErrorContains(t, err, registry.ErrInvalidRegistry.Error())
}

func (ts *RegistryServiceTestSuite) TestGet_Success() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	expectedRegistry := assertCreateRegistry(t)

	ts.repository.EXPECT().
		Get(mock.Anything, organizationID, expectedRegistry.ID.String()).
		Return(expectedRegistry, nil)

	reg, err := ts.service.Get(context.Background(), organizationID, expectedRegistry.ID.String())
	assert.Nil(t, err)
	assertEqualRegistry(t, expectedRegistry, reg)
}
func (ts *RegistryServiceTestSuite) TestUpdate_FailWithInvalidOrganizationID() {
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
			reg, err := ts.service.Update(context.Background(), tc.OrganizationID, gofakeit.UUID(), assertNewRegistryUpsertRequest(t))
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToUpdateRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidRegistryOrganizationIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestUpdate_FailWithInvalidRegistryID() {
	t := ts.T()

	testCases := []struct {
		TestName   string
		RegistryID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:   "invalid_uuid",
			RegistryID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			reg, err := ts.service.Update(context.Background(), gofakeit.UUID(), tc.RegistryID, assertNewRegistryUpsertRequest(t))
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToUpdateRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidRegistryIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestUpdate_FailRegistryNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, fakeID, mock.Anything).
		Return(nil, registry.ErrInvalidRegistry)

	reg, err := ts.service.Update(context.Background(), fakeID, fakeID, assertNewRegistryUpsertRequest(t))
	assert.Nil(t, reg)
	assert.ErrorContains(t, err, registry.ErrFailedToUpdateRegistry.Error())
	assert.ErrorContains(t, err, registry.ErrInvalidRegistry.Error())
}

func (ts *RegistryServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest registry.UpsertRequest
	}{
		{
			TestName: "invalid_name",
			UpdateRequest: registry.UpsertRequest{
				Kind: registry.KindDockerHub.String(),
				URL:  gofakeit.URL(),
			},
		},
		{
			TestName: "invalid_kind",
			UpdateRequest: registry.UpsertRequest{
				Name: gofakeit.Name(),
				URL:  gofakeit.URL(),
			},
		},
		{
			TestName: "invalid_url",
			UpdateRequest: registry.UpsertRequest{
				Name: gofakeit.Name(),
				Kind: registry.KindDockerHub.String(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			reg, err := ts.service.Update(context.Background(), gofakeit.UUID(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, reg)
			assert.ErrorContains(t, err, registry.ErrFailedToUpdateRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	expectedRegistry := assertCreateRegistry(t)

	ts.repository.EXPECT().
		Update(mock.Anything, organizationID, expectedRegistry.ID.String(), mock.Anything).
		Return(expectedRegistry, nil)

	reg, err := ts.service.Update(context.Background(), organizationID, expectedRegistry.ID.String(), assertNewRegistryUpsertRequest(t))
	assert.Nil(t, err)
	assertEqualRegistry(t, expectedRegistry, reg)
}

func (ts *RegistryServiceTestSuite) TestDelete_FailWithInvalidOrganizationID() {
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
			err := ts.service.Delete(context.Background(), tc.OrganizationID, gofakeit.UUID())
			assert.ErrorContains(t, err, registry.ErrFailedToDeleteRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestDelete_FailWithInvalidRegistryID() {
	t := ts.T()

	testCases := []struct {
		TestName   string
		RegistryID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:   "invalid_uuid",
			RegistryID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.service.Delete(context.Background(), gofakeit.UUID(), tc.RegistryID)
			assert.ErrorContains(t, err, registry.ErrFailedToDeleteRegistry.Error())
			assert.ErrorContains(t, err, registry.ErrInvalidRegistryIDParam.Error())
		})
	}
}

func (ts *RegistryServiceTestSuite) TestDelete_FailRegistryNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID, fakeID).
		Return(registry.ErrInvalidRegistry)

	err := ts.service.Delete(context.Background(), fakeID, fakeID)
	assert.ErrorContains(t, err, registry.ErrFailedToDeleteRegistry.Error())
	assert.ErrorContains(t, err, registry.ErrInvalidRegistry.Error())
}

func (ts *RegistryServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	expectedRegistry := assertCreateRegistry(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, organizationID, expectedRegistry.ID.String()).
		Return(nil)

	err := ts.service.Delete(context.Background(), organizationID, expectedRegistry.ID.String())
	assert.Nil(t, err)
}

func TestRegistryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryServiceTestSuite))
}

func assertCreateRegistry(t *testing.T) *registry.Registry {
	kindIdx := gofakeit.IntRange(0, len(registry.AllowedKindValues)-1)

	reg, err := registry.NewRegistry(registry.NewRegistryParams{
		RegistryID:     gofakeit.UUID(),
		OrganizationID: gofakeit.UUID(),
		Name:           gofakeit.Name(),
		Kind:           registry.AllowedKindValues[kindIdx].String(),
		URL:            gofakeit.URL(),
		Description:    pointer.ToString(gofakeit.Word()),
	})
	require.NoError(t, err)
	require.NotNil(t, reg)
	require.NoError(t, reg.Validate())

	return reg
}

func assertNewRegistryUpsertRequest(t *testing.T) registry.UpsertRequest {
	kindIdx := gofakeit.IntRange(0, len(registry.AllowedKindValues)-1)

	req := registry.UpsertRequest{
		Name:        gofakeit.Name(),
		Kind:        registry.AllowedKindValues[kindIdx].String(),
		URL:         gofakeit.URL(),
		Description: pointer.ToString(gofakeit.Word()),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertEqualRegistry(t *testing.T, expected *registry.Registry, actual *registry.Registry) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.OrganizationID, actual.OrganizationID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Kind, actual.Kind)
	assert.Equal(t, expected.URL.String(), actual.URL.String())
	assert.Equal(t, expected.Description, actual.Description)
	assertEqualRegistryConfig(t, expected.Config, actual.Config)
}

func assertEqualRegistryConfig(t *testing.T, expected map[string]string, actual map[string]string) {
	assert.Len(t, expected, len(actual))
	for k, v := range expected {
		assert.Equal(t, v, actual[k])
	}
}
