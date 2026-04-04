//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

func TestNewQoveryEnvSecretVariableRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     secret.UpsertRequest
		ParentId    string
		ParentScope qovery.APIVariableScopeEnum
	}{
		{
			TestName: "success",
			Request: secret.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
			ParentId:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
		{
			TestName: "success_with_mount_path",
			Request: secret.UpsertRequest{
				Key:       gofakeit.Word(),
				Value:     gofakeit.Word(),
				MountPath: "/usr/local/secrets/api-key",
			},
			ParentId:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
		},
		{
			TestName: "success_without_mount_path",
			Request: secret.UpsertRequest{
				Key:   gofakeit.Word(),
				Value: gofakeit.Word(),
			},
			ParentId:    gofakeit.UUID(),
			ParentScope: qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := newQoveryEnvSecretVariableRequestFromDomain(tc.Request, tc.ParentId, tc.ParentScope)
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.Request.Value, req.Value)
			assert.True(t, req.IsSecret)
			assert.Equal(t, tc.ParentId, req.VariableParentId)
			assert.Equal(t, tc.ParentScope, req.VariableScope)
			if tc.Request.MountPath != "" {
				assert.True(t, req.MountPath.IsSet())
				assert.Equal(t, tc.Request.MountPath, *req.MountPath.Get())
			} else {
				assert.False(t, req.MountPath.IsSet())
			}
		})
	}
}

func TestNewDomainEnvSecretFromQovery(t *testing.T) {
	t.Parallel()

	now := time.Now()
	description := "test description"
	mountPathValue := "/usr/local/secrets/api-key"

	testCases := []struct {
		TestName      string
		Variable      *qovery.VariableResponse
		ExpectedError bool
	}{
		{
			TestName:      "error_nil_variable",
			Variable:      nil,
			ExpectedError: true,
		},
		{
			TestName: "success",
			Variable: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Key:          gofakeit.Word(),
				Value:        *qovery.NewNullableString(func() *string { v := gofakeit.Word(); return &v }()),
				Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
				VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
				CreatedAt:    now,
				IsSecret:     true,
				Description:  &description,
			},
		},
		{
			TestName: "success_with_mount_path",
			Variable: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Key:          gofakeit.Word(),
				Value:        *qovery.NewNullableString(func() *string { v := "secret-file-content"; return &v }()),
				MountPath:    *qovery.NewNullableString(&mountPathValue),
				Scope:        qovery.APIVARIABLESCOPEENUM_APPLICATION,
				VariableType: qovery.APIVARIABLETYPEENUM_FILE,
				CreatedAt:    now,
				IsSecret:     true,
				Description:  &description,
			},
		},
		{
			TestName: "success_with_nil_mount_path",
			Variable: &qovery.VariableResponse{
				Id:           gofakeit.UUID(),
				Key:          gofakeit.Word(),
				Value:        *qovery.NewNullableString(func() *string { v := gofakeit.Word(); return &v }()),
				Scope:        qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
				VariableType: qovery.APIVARIABLETYPEENUM_VALUE,
				CreatedAt:    now,
				IsSecret:     true,
				Description:  &description,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			v, err := newDomainEnvSecretFromQovery(tc.Variable)
			if tc.ExpectedError {
				assert.Error(t, err)
				assert.Nil(t, v)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, v)
			assert.Equal(t, tc.Variable.Id, v.ID.String())
			assert.Equal(t, tc.Variable.Key, v.Key)
			assert.Equal(t, string(tc.Variable.Scope), v.Scope.String())
			assert.Equal(t, string(tc.Variable.VariableType), v.Type)

			expectedMountPath := ""
			if tc.Variable.MountPath.IsSet() && tc.Variable.MountPath.Get() != nil {
				expectedMountPath = *tc.Variable.MountPath.Get()
			}
			assert.Equal(t, expectedMountPath, v.MountPath)
		})
	}
}
