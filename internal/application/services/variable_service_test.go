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
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

const (
	// Define constant min / max range when faking arrays.
	arraySizeRangeMin = 1
	arraySizeRangeMax = 5
)

type VariableServiceTestSuite struct {
	suite.Suite

	repository *mocks_test.VariableRepository
	service    variable.Service
}

func (ts *VariableServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize repository
	variableRepository := mocks_test.NewVariableRepository(t)

	// Initialize service
	variableService, err := services.NewVariableService(variableRepository)
	require.NoError(t, err)
	require.NotNil(t, variableService)

	ts.repository = variableRepository
	ts.service = variableService
}

func (ts *VariableServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	variableService, err := services.NewVariableService(nil)
	assert.Nil(t, variableService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *VariableServiceTestSuite) TestNew_Success() {
	t := ts.T()

	variableService, err := services.NewVariableService(mocks_test.NewVariableRepository(t))
	assert.Nil(t, err)
	assert.NotNil(t, variableService)
}

func (ts *VariableServiceTestSuite) TestList_FailWithInvalidResourceID() {
	t := ts.T()

	testCases := []struct {
		TestName   string
		VariableID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:   "invalid_uuid",
			VariableID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			vars, err := ts.service.List(context.Background(), tc.VariableID)
			assert.Nil(t, vars)
			assert.ErrorContains(t, err, variable.ErrFailedToListVariables.Error())
			assert.ErrorContains(t, err, variable.ErrInvalidResourceIDParam.Error())
		})
	}
}

func (ts *VariableServiceTestSuite) TestList_FailWithMissingVariable() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		List(mock.Anything, fakeID).
		Return(nil, errors.New(""))

	vars, err := ts.service.List(context.Background(), fakeID)
	assert.Nil(t, vars)
	assert.ErrorContains(t, err, variable.ErrFailedToListVariables.Error())
}

func (ts *VariableServiceTestSuite) TestList_Success() {
	t := ts.T()

	resourceID := gofakeit.UUID()
	expectedVars := assertCreateVariables(t)

	ts.repository.EXPECT().
		List(mock.Anything, resourceID).
		Return(expectedVars, nil)

	vars, err := ts.service.List(context.Background(), resourceID)
	assert.Nil(t, err)
	assertEqualVariables(t, expectedVars, vars)
}

func (ts *VariableServiceTestSuite) TestUpdate_FailWithInvalidResourceID() {
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
			vars, err := ts.service.Update(context.Background(), tc.ResourceID, assertNewVariableDiffRequest(t))
			assert.Nil(t, vars)
			assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
			assert.ErrorContains(t, err, variable.ErrInvalidResourceIDParam.Error())
		})
	}
}

func (ts *VariableServiceTestSuite) TestUpdate_FailResourceNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewVariableDiffRequest(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID, mock.Anything).
		Return(errors.New(""))

	vars, err := ts.service.Update(context.Background(), fakeID, updateRequest)
	assert.Nil(t, vars)
	assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
}

func (ts *VariableServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName    string
		DiffRequest variable.DiffRequest
	}{
		{
			TestName: "invalid_create",
			DiffRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{},
					},
				},
			},
		},
		{
			TestName: "invalid_update",
			DiffRequest: variable.DiffRequest{
				Update: []variable.DiffUpdateRequest{
					{
						UpsertRequest: assertNewVariableUpsertRequest(t),
					},
				},
			},
		},
		{
			TestName: "invalid_delete",
			DiffRequest: variable.DiffRequest{
				Delete: []variable.DiffDeleteRequest{
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
			assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
			assert.ErrorContains(t, err, variable.ErrInvalidDiffRequest.Error())
		})
	}
}

func (ts *VariableServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	resourceID := gofakeit.UUID()
	updateRequest := assertNewVariableDiffRequest(t)

	expectedVars := make(variable.Variables, 0, len(updateRequest.Create)+len(updateRequest.Update))
	for _, toCreate := range updateRequest.Create {
		v := assertCreateVariableFromDiffCreateRequest(t, toCreate)
		expectedVars = append(expectedVars, *v)

		ts.repository.EXPECT().
			Create(mock.Anything, resourceID, toCreate.UpsertRequest).
			Return(v, nil)
	}

	for _, toUpdate := range updateRequest.Update {
		v := assertCreateVariableFromDiffUpdateRequest(t, toUpdate)
		expectedVars = append(expectedVars, *v)

		ts.repository.EXPECT().
			Update(mock.Anything, resourceID, toUpdate.VariableID, toUpdate.UpsertRequest).
			Return(v, nil)
	}

	ts.repository.EXPECT().
		Delete(mock.Anything, resourceID, mock.Anything).
		Return(nil)

	vars, err := ts.service.Update(context.Background(), resourceID, updateRequest)
	assert.Nil(t, err)
	assertEqualVariables(t, expectedVars, vars)
}

func TestVariableServiceTestSuite(t *testing.T) {
	suite.Run(t, new(VariableServiceTestSuite))
}

func assertNewVariableDiffRequest(t *testing.T) variable.DiffRequest {
	size := gofakeit.IntRange(arraySizeRangeMin, arraySizeRangeMax)

	req := variable.DiffRequest{
		Create: make([]variable.DiffCreateRequest, 0, size),
		Update: make([]variable.DiffUpdateRequest, 0, size),
		Delete: make([]variable.DiffDeleteRequest, 0, size),
	}

	for i := 0; i < size; i++ {
		req.Create = append(req.Create, variable.DiffCreateRequest{
			UpsertRequest: assertNewVariableUpsertRequest(t),
		})

		req.Update = append(req.Update, variable.DiffUpdateRequest{
			UpsertRequest: assertNewVariableUpsertRequest(t),
			VariableID:    gofakeit.UUID(),
		})

		req.Delete = append(req.Delete, variable.DiffDeleteRequest{
			VariableID: gofakeit.UUID(),
		})
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewVariableUpsertRequest(t *testing.T) variable.UpsertRequest {
	req := variable.UpsertRequest{
		Key:   gofakeit.Word(),
		Value: gofakeit.Word(),
	}

	require.NoError(t, req.Validate())

	return req
}

func assertCreateVariables(t *testing.T) variable.Variables {
	size := gofakeit.IntRange(arraySizeRangeMin, arraySizeRangeMax)

	vars := make(variable.Variables, 0, size)
	for i := 0; i < size; i++ {
		vars = append(vars, *assertCreateVariable(t))
	}
	require.Len(t, vars, size)

	return vars
}

func assertCreateVariable(t *testing.T) *variable.Variable {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := variable.NewVariable(variable.NewVariableParams{
		VariableID: gofakeit.UUID(),
		Scope:      variable.AllowedScopeValues[scopeIdx].String(),
		Key:        gofakeit.Word(),
		Value:      gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NoError(t, v.Validate())

	return v
}

func assertCreateVariablesFromDiffRequest(t *testing.T, req variable.DiffRequest) variable.Variables {
	created := assertCreateVariablesFromDiffCreateRequest(t, req.Create)
	updated := assertCreateVariablesFromDiffUpdateRequest(t, req.Update)

	return append(created, updated...)
}

func assertCreateVariablesFromDiffCreateRequest(t *testing.T, req []variable.DiffCreateRequest) variable.Variables {
	vars := make(variable.Variables, 0, len(req))
	for _, r := range req {
		vars = append(vars, *assertCreateVariableFromDiffCreateRequest(t, r))
	}

	return vars
}

func assertCreateVariableFromDiffCreateRequest(t *testing.T, req variable.DiffCreateRequest) *variable.Variable {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := variable.NewVariable(variable.NewVariableParams{
		VariableID: gofakeit.UUID(),
		Scope:      variable.AllowedScopeValues[scopeIdx].String(),
		Key:        req.Key,
		Value:      req.Value,
	})
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NoError(t, v.Validate())

	return v
}

func assertCreateVariablesFromDiffUpdateRequest(t *testing.T, req []variable.DiffUpdateRequest) variable.Variables {
	vars := make(variable.Variables, 0, len(req))
	for _, r := range req {
		vars = append(vars, *assertCreateVariableFromDiffUpdateRequest(t, r))
	}

	return vars
}
func assertCreateVariableFromDiffUpdateRequest(t *testing.T, req variable.DiffUpdateRequest) *variable.Variable {
	scopeIdx := gofakeit.IntRange(0, len(variable.AllowedScopeValues)-1)

	v, err := variable.NewVariable(variable.NewVariableParams{
		VariableID: req.VariableID,
		Scope:      variable.AllowedScopeValues[scopeIdx].String(),
		Key:        req.Key,
		Value:      req.Value,
	})
	require.NoError(t, err)
	require.NotNil(t, v)

	require.NoError(t, v.Validate())

	return v
}

func assertEqualVariables(t *testing.T, expected variable.Variables, actual variable.Variables) {
	require.Len(t, expected, len(actual))

	actualByID := map[string]variable.Variable{}
	for _, v := range actual {
		actualByID[v.ID.String()] = v
	}

	for _, v := range expected {
		found, ok := actualByID[v.ID.String()]
		require.True(t, ok)
		assertEqualVariable(t, &v, &found)
	}
}

func assertEqualVariable(t *testing.T, expected *variable.Variable, actual *variable.Variable) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Scope, actual.Scope)
	assert.Equal(t, expected.Key, actual.Key)
	assert.Equal(t, expected.Value, actual.Value)
}
