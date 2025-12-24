package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewQoveryEnvVariableRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     variable.UpsertRequest
		IsSecret    bool
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success_not_secret",
			Request: variable.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			IsSecret:    false,
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
		{
			TestName: "success_secret",
			Request: variable.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			IsSecret:    true,
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvVariableRequestFromDomain(tc.Request, tc.IsSecret, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.Request.Value, req.Value)
			assert.Equal(t, tc.IsSecret, req.IsSecret)
			assert.Equal(t, tc.ParentScope, req.VariableScope)
			assert.Equal(t, tc.ParentID, req.VariableParentId)
		})
	}
}

func TestNewQoveryEnvVariableCreateAliasRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     variable.UpsertRequest
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: variable.UpsertRequest{
				Key:         gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvVariableCreateAliasRequestFromDomain(tc.Request, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.ParentScope, req.AliasScope)
			assert.Equal(t, tc.ParentID, req.AliasParentId)
		})
	}
}

func TestNewQoveryEnvVariableCreateOverrideRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     variable.UpsertRequest
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: variable.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvVariableCreateOverrideRequestFromDomain(tc.Request, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Value, req.Value)
			assert.Equal(t, tc.ParentScope, req.OverrideScope)
			assert.Equal(t, tc.ParentID, req.OverrideParentId)
		})
	}
}

func TestNewDomainEnvVariablesFromQovery(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)

	testCases := []struct {
		TestName      string
		List          *qovery.VariableResponseList
		ExpectedLen   int
		ExpectedError bool
	}{
		{
			TestName:    "success_with_empty_list",
			List:        &qovery.VariableResponseList{Results: []qovery.VariableResponse{}},
			ExpectedLen: 0,
		},
		{
			TestName: "success_with_variables",
			List: &qovery.VariableResponseList{
				Results: []qovery.VariableResponse{
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
						Key:          gofakeit.Word(),
						Value:        *qovery.NewNullableString(func() *string { s := gofakeit.Word(); return &s }()),
						VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
						Description:  &description,
					},
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
						Key:          gofakeit.Word(),
						Value:        *qovery.NewNullableString(func() *string { s := gofakeit.Word(); return &s }()),
						VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
						Description:  &description,
					},
				},
			},
			ExpectedLen: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			vars, err := newDomainEnvVariablesFromQovery(tc.List)
			if tc.ExpectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, vars, tc.ExpectedLen)

			for idx, v := range vars {
				assert.True(t, v.IsValid())
				assert.Equal(t, tc.List.Results[idx].Id, v.ID.String())
				assert.Equal(t, string(tc.List.Results[idx].Scope), v.Scope.String())
				assert.Equal(t, tc.List.Results[idx].Key, v.Key)
			}
		})
	}
}

func TestNewDomainEnvVariableFromQovery(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)
	value := gofakeit.Word()

	testCases := []struct {
		TestName      string
		Input         *qovery.VariableResponse
		ExpectedError error
	}{
		{
			TestName:      "error_with_nil_input",
			Input:         nil,
			ExpectedError: variable.ErrNilVariable,
		},
		{
			TestName: "success",
			Input: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
				Key:          gofakeit.Word(),
				Value:        *qovery.NewNullableString(&value),
				VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
				Description:  &description,
			},
		},
		{
			TestName: "success_with_nil_value",
			Input: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
				Key:          gofakeit.Word(),
				Value:        qovery.NullableString{},
				VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
				Description:  &description,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newDomainEnvVariableFromQovery(tc.Input)
			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsValid())
			assert.Equal(t, tc.Input.Id, result.ID.String())
			assert.Equal(t, string(tc.Input.Scope), result.Scope.String())
			assert.Equal(t, tc.Input.Key, result.Key)
		})
	}
}

func TestNewQoveryEnvVariableEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  variable.UpsertRequest
	}{
		{
			TestName: "success",
			Request: variable.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvVariableEditRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
		})
	}
}
