package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"

)

func TestNewDomainVariablesFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Variables     *qovery.EnvironmentVariableResponseList
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Variables: &qovery.EnvironmentVariableResponseList{
				Results: []qovery.EnvironmentVariable{
					{
						Id:    gofakeit.UUID(),
						Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
						Key:   gofakeit.Word(),
						Value: func() *string {
							v := gofakeit.Word()
							return &v
						}(),
					},
					{
						Id:    gofakeit.UUID(),
						Scope: qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
						Key:   gofakeit.Word(),
						Value: func() *string {
							v := gofakeit.Word()
							return &v
						}(),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			vv, err := newDomainVariablesFromQovery(tc.Variables)
			assert.NoError(t, err)
			assert.Len(t, tc.Variables.GetResults(), len(vv))

			for idx, v := range vv {
				assert.True(t, v.IsValid())
				assert.Equal(t, tc.Variables.GetResults()[idx].Id, v.ID.String())
				assert.Equal(t, string(tc.Variables.GetResults()[idx].Scope), v.Scope.String())
				assert.Equal(t, tc.Variables.GetResults()[idx].Key, v.Key)
				value := ""
				if tc.Variables.GetResults()[idx].Value != nil {
					value = *tc.Variables.GetResults()[idx].Value
				}
				assert.Equal(t, value, v.Value)
			}
		})
	}
}

func TestNewQoveryEnvironmentVariableRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       variable.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success",
			Request: variable.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryEnvironmentVariableRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
			value := ""
			if req.Value != nil {
				value = *req.Value
			}
			assert.Equal(t, tc.Request.Value, value)
		})
	}
}

func TestNewQoveryEnvironmentVariableEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       variable.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success",
			Request: variable.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryEnvironmentVariableEditRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.Request.Value, req.Value)
		})
	}
}
