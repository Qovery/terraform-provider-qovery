package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

func TestNewQoveryEnvSecretVariableRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     secret.UpsertRequest
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
		{
			TestName: "success_with_environment_scope",
			Request: secret.UpsertRequest{
				Key:         gofakeit.Word(),
				Value:       gofakeit.Word(),
				Description: gofakeit.Sentence(5),
			},
			ParentID:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvSecretVariableRequestFromDomain(tc.Request, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.Request.Value, req.Value)
			assert.True(t, req.IsSecret)
			assert.Equal(t, tc.ParentScope, req.VariableScope)
			assert.Equal(t, tc.ParentID, req.VariableParentId)
		})
	}
}

func TestNewDomainEnvSecretsFromQovery(t *testing.T) {
	t.Parallel()

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
			TestName: "success_with_secrets",
			List: &qovery.VariableResponseList{
				Results: []qovery.VariableResponse{
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
						Key:          gofakeit.Word(),
						VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
						Description:  func() *string { s := gofakeit.Sentence(3); return &s }(),
					},
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
						Key:          gofakeit.Word(),
						VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
						Description:  func() *string { s := gofakeit.Sentence(3); return &s }(),
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
			secrets, err := newDomainEnvSecretsFromQovery(tc.List)
			if tc.ExpectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, secrets, tc.ExpectedLen)

			for idx, s := range secrets {
				assert.True(t, s.IsValid())
				assert.Equal(t, tc.List.Results[idx].Id, s.ID.String())
				assert.Equal(t, string(tc.List.Results[idx].Scope), s.Scope.String())
				assert.Equal(t, tc.List.Results[idx].Key, s.Key)
			}
		})
	}
}

func TestNewDomainEnvSecretFromQovery(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)

	testCases := []struct {
		TestName      string
		Input         *qovery.VariableResponse
		ExpectedError error
	}{
		{
			TestName:      "error_with_nil_input",
			Input:         nil,
			ExpectedError: secret.ErrNilSecret,
		},
		{
			TestName: "success",
			Input: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
				Key:          gofakeit.Word(),
				VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
				Description:  &description,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newDomainEnvSecretFromQovery(tc.Input)
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

func TestNewQoveryEnvSecretEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  secret.UpsertRequest
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
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
			req := newQoveryEnvSecretEditRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
		})
	}
}

func TestNewQoveryEnvSecretCreateAliasRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     secret.UpsertRequest
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
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
			req := newQoveryEnvSecretCreateAliasRequestFromDomain(tc.Request, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.ParentScope, req.AliasScope)
			assert.Equal(t, tc.ParentID, req.AliasParentId)
		})
	}
}

func TestNewQoveryEnvSecretCreateOverrideRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     secret.UpsertRequest
		ParentID    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
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
			req := newQoveryEnvSecretCreateOverrideRequestFromDomain(tc.Request, tc.ParentID, tc.ParentScope)
			assert.Equal(t, tc.Request.Value, req.Value)
			assert.Equal(t, tc.ParentScope, req.OverrideScope)
			assert.Equal(t, tc.ParentID, req.OverrideParentId)
		})
	}
}
