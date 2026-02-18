package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

func TestNewDomainSecretsFromQovery(t *testing.T) {
	t.Parallel()

	variableType := qovery.APIVARIABLETYPEENUM_VALUE
	testCases := []struct {
		TestName      string
		Secrets       *qovery.SecretResponseList
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Secrets: &qovery.SecretResponseList{
				Results: []qovery.Secret{
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
						Key:          gofakeit.Word(),
						VariableType: &variableType,
					},
					{
						Id:           gofakeit.UUID(),
						Scope:        qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
						Key:          gofakeit.Word(),
						VariableType: &variableType,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			vv, err := newDomainSecretsFromQovery(tc.Secrets)
			assert.NoError(t, err)
			assert.Len(t, tc.Secrets.GetResults(), len(vv))

			for idx, v := range vv {
				assert.True(t, v.IsValid())
				assert.Equal(t, tc.Secrets.GetResults()[idx].Id, v.ID.String())
				assert.Equal(t, string(tc.Secrets.GetResults()[idx].Scope), v.Scope.String())
				assert.Equal(t, tc.Secrets.GetResults()[idx].Key, v.Key)
				assert.Equal(t, tc.Secrets.GetResults()[idx].VariableType, &variableType)
			}
		})
	}
}

func TestNewQoveryEnvironmentSecretRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       secret.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoverySecretRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
			value := ""
			if req.Value != nil {
				value = *req.Value
			}
			assert.Equal(t, tc.Request.Value, value)
		})
	}
}

func TestNewQoveryEnvironmentSecretEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       secret.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoverySecretEditRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
			value := ""
			if req.Value != nil {
				value = *req.Value
			}
			assert.Equal(t, tc.Request.Value, value)
		})
	}
}
