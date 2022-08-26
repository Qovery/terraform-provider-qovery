//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type SecretServiceTestSuite struct {
	suite.Suite

	repository *mocks_test.SecretRepository
	service    secret.Service
}

func (ts *SecretServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize repository
	secretRepository := mocks_test.NewSecretRepository(t)

	// Initialize service
	secretService, err := services.NewSecretService(secretRepository)
	require.NoError(t, err)
	require.NotNil(t, secretService)

	ts.repository = secretRepository
	ts.service = secretService
}

func (ts *SecretServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	secretService, err := services.NewSecretService(nil)
	assert.Nil(t, secretService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *SecretServiceTestSuite) TestNew_Success() {
	t := ts.T()

	secretService, err := services.NewSecretService(mocks_test.NewSecretRepository(t))
	assert.Nil(t, err)
	assert.NotNil(t, secretService)
}

func (ts *SecretServiceTestSuite) TestList_FailWithInvalidResourceID() {
	t := ts.T()

	testCases := []struct {
		TestName string
		SecretID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName: "invalid_uuid",
			SecretID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			vars, err := ts.service.List(context.Background(), tc.SecretID)
			assert.Nil(t, vars)
			assert.ErrorContains(t, err, secret.ErrFailedToListSecrets.Error())
			assert.ErrorContains(t, err, secret.ErrInvalidResourceIDParam.Error())
		})
	}
}

func (ts *SecretServiceTestSuite) TestList_FailWithMissingSecret() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		List(mock.Anything, fakeID).
		Return(nil, errors.New(""))

	vars, err := ts.service.List(context.Background(), fakeID)
	assert.Nil(t, vars)
	assert.ErrorContains(t, err, secret.ErrFailedToListSecrets.Error())
}

func (ts *SecretServiceTestSuite) TestList_Success() {
	t := ts.T()

	resourceID := gofakeit.UUID()
	expectedVars := assertCreateSecrets(t)

	ts.repository.EXPECT().
		List(mock.Anything, resourceID).
		Return(expectedVars, nil)

	vars, err := ts.service.List(context.Background(), resourceID)
	assert.Nil(t, err)
	assertEqualSecrets(t, expectedVars, vars)
}

func (ts *SecretServiceTestSuite) TestUpdate_FailWithInvalidResourceID() {
	t := ts.T()

	testCases := []struct {
		TestName   string
		ResourceID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:   "invalid_uuid",
			ResourceID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			vars, err := ts.service.Update(context.Background(), tc.ResourceID, assertNewSecretDiffRequest(t))
			assert.Nil(t, vars)
			assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
			assert.ErrorContains(t, err, secret.ErrInvalidResourceIDParam.Error())
		})
	}
}

func (ts *SecretServiceTestSuite) TestUpdate_FailResourceNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewSecretDiffRequest(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID, mock.Anything).
		Return(errors.New(""))

	vars, err := ts.service.Update(context.Background(), fakeID, updateRequest)
	assert.Nil(t, vars)
	assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
}

func (ts *SecretServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName    string
		DiffRequest secret.DiffRequest
	}{
		{
			TestName: "invalid_create",
			DiffRequest: secret.DiffRequest{
				Create: []secret.DiffCreateRequest{
					{
						UpsertRequest: secret.UpsertRequest{},
					},
				},
			},
		},
		{
			TestName: "invalid_update",
			DiffRequest: secret.DiffRequest{
				Update: []secret.DiffUpdateRequest{
					{
						UpsertRequest: assertNewSecretUpsertRequest(t),
					},
				},
			},
		},
		{
			TestName: "invalid_delete",
			DiffRequest: secret.DiffRequest{
				Delete: []secret.DiffDeleteRequest{
					{},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			vars, err := ts.service.Update(context.Background(), gofakeit.UUID(), tc.DiffRequest)
			assert.Nil(t, vars)
			assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
			assert.ErrorContains(t, err, secret.ErrInvalidDiffRequest.Error())
		})
	}
}

func (ts *SecretServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	resourceID := gofakeit.UUID()
	updateRequest := assertNewSecretDiffRequest(t)

	expectedVars := make(secret.Secrets, 0, len(updateRequest.Create)+len(updateRequest.Update))
	for _, toCreate := range updateRequest.Create {
		v := assertCreateSecretFromDiffCreateRequest(t, toCreate)
		expectedVars = append(expectedVars, *v)

		ts.repository.EXPECT().
			Create(mock.Anything, resourceID, toCreate.UpsertRequest).
			Return(v, nil)
	}

	for _, toUpdate := range updateRequest.Update {
		v := assertCreateSecretFromDiffUpdateRequest(t, toUpdate)
		expectedVars = append(expectedVars, *v)

		ts.repository.EXPECT().
			Update(mock.Anything, resourceID, toUpdate.SecretID, toUpdate.UpsertRequest).
			Return(v, nil)
	}

	ts.repository.EXPECT().
		Delete(mock.Anything, resourceID, mock.Anything).
		Return(nil)

	vars, err := ts.service.Update(context.Background(), resourceID, updateRequest)
	assert.Nil(t, err)
	assertEqualSecrets(t, expectedVars, vars)
}

func TestSecretServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SecretServiceTestSuite))
}

func assertNewSecretDiffRequest(t *testing.T) secret.DiffRequest {
	size := gofakeit.IntRange(arraySizeRangeMin, arraySizeRangeMax)

	req := secret.DiffRequest{
		Create: make([]secret.DiffCreateRequest, 0, size),
		Update: make([]secret.DiffUpdateRequest, 0, size),
		Delete: make([]secret.DiffDeleteRequest, 0, size),
	}

	for i := 0; i < size; i++ {
		req.Create = append(req.Create, secret.DiffCreateRequest{
			UpsertRequest: assertNewSecretUpsertRequest(t),
		})

		req.Update = append(req.Update, secret.DiffUpdateRequest{
			UpsertRequest: assertNewSecretUpsertRequest(t),
			SecretID:      gofakeit.UUID(),
		})

		req.Delete = append(req.Delete, secret.DiffDeleteRequest{
			SecretID: gofakeit.UUID(),
		})
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewSecretUpsertRequest(t *testing.T) secret.UpsertRequest {
	req := secret.UpsertRequest{
		Key:   gofakeit.Word(),
		Value: gofakeit.Word(),
	}

	require.NoError(t, req.Validate())

	return req
}

func assertCreateSecrets(t *testing.T) secret.Secrets {
	size := gofakeit.IntRange(arraySizeRangeMin, arraySizeRangeMax)

	vars := make(secret.Secrets, 0, size)
	for i := 0; i < size; i++ {
		vars = append(vars, *assertCreateSecret(t))
	}
	require.Len(t, vars, size)

	return vars
}

func assertCreateSecret(t *testing.T) *secret.Secret {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := secret.NewSecret(secret.NewSecretParams{
		SecretID: gofakeit.UUID(),
		Scope:    variable.AllowedScopeValues[scopeIdx].String(),
		Key:      gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NoError(t, v.Validate())

	return v
}

func assertCreateSecretsFromDiffRequest(t *testing.T, req secret.DiffRequest) secret.Secrets {
	created := assertCreateSecretsFromDiffCreateRequest(t, req.Create)
	updated := assertCreateSecretsFromDiffUpdateRequest(t, req.Update)

	return append(created, updated...)
}

func assertCreateSecretsFromDiffCreateRequest(t *testing.T, req []secret.DiffCreateRequest) secret.Secrets {
	vars := make(secret.Secrets, 0, len(req))
	for _, r := range req {
		vars = append(vars, *assertCreateSecretFromDiffCreateRequest(t, r))
	}

	return vars
}

func assertCreateSecretFromDiffCreateRequest(t *testing.T, req secret.DiffCreateRequest) *secret.Secret {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := secret.NewSecret(secret.NewSecretParams{
		SecretID: gofakeit.UUID(),
		Scope:    variable.AllowedScopeValues[scopeIdx].String(),
		Key:      req.Key,
	})
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NoError(t, v.Validate())

	return v
}

func assertCreateSecretsFromDiffUpdateRequest(t *testing.T, req []secret.DiffUpdateRequest) secret.Secrets {
	vars := make(secret.Secrets, 0, len(req))
	for _, r := range req {
		vars = append(vars, *assertCreateSecretFromDiffUpdateRequest(t, r))
	}

	return vars
}

func assertCreateSecretFromDiffUpdateRequest(t *testing.T, req secret.DiffUpdateRequest) *secret.Secret {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := secret.NewSecret(secret.NewSecretParams{
		SecretID: req.SecretID,
		Scope:    variable.AllowedScopeValues[scopeIdx].String(),
		Key:      req.Key,
	})
	require.NoError(t, err)
	require.NotNil(t, v)

	require.NoError(t, v.Validate())

	return v
}

func assertEqualSecrets(t *testing.T, expected secret.Secrets, actual secret.Secrets) {
	require.Len(t, expected, len(actual))

	actualByID := map[string]secret.Secret{}
	for _, v := range actual {
		actualByID[v.ID.String()] = v
	}

	for _, v := range expected {
		found, ok := actualByID[v.ID.String()]
		require.True(t, ok)
		assertEqualSecret(t, &v, &found)
	}
}

func assertEqualSecret(t *testing.T, expected *secret.Secret, actual *secret.Secret) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Scope, actual.Scope)
	assert.Equal(t, expected.Key, actual.Key)
}
