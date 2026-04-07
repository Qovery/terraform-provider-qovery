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

	variableType := qovery.APIVARIABLETYPEENUM_VALUE
	fileVariableType := qovery.APIVARIABLETYPEENUM_FILE
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
						VariableType: variableType,
					},
					{
						Id:    gofakeit.UUID(),
						Scope: qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
						Key:   gofakeit.Word(),
						Value: func() *string {
							v := gofakeit.Word()
							return &v
						}(),
						VariableType: variableType,
					},
				},
			},
		},
		{
			TestName: "success_with_file_type",
			Variables: &qovery.EnvironmentVariableResponseList{
				Results: []qovery.EnvironmentVariable{
					{
						Id:    gofakeit.UUID(),
						Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION,
						Key:   gofakeit.Word(),
						Value: func() *string {
							v := "file-content"
							return &v
						}(),
						MountPath: *qovery.NewNullableString(func() *string {
							v := "/etc/app/config.yaml"
							return &v
						}()),
						VariableType: fileVariableType,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
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
				assert.Equal(t, string(tc.Variables.GetResults()[idx].VariableType), v.Type)
				mountPath := ""
				if tc.Variables.GetResults()[idx].MountPath.IsSet() && tc.Variables.GetResults()[idx].MountPath.Get() != nil {
					mountPath = *tc.Variables.GetResults()[idx].MountPath.Get()
				}
				assert.Equal(t, mountPath, v.MountPath)
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
		{
			TestName: "success_with_mount_path",
			Request: variable.UpsertRequest{
				Key:       gofakeit.Word(),
				Value:     gofakeit.Word(),
				MountPath: "/etc/config/app.yaml",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryEnvironmentVariableRequestFromDomain(tc.Request)
			assert.Equal(t, tc.Request.Key, req.Key)
			value := ""
			if req.Value != nil {
				value = *req.Value
			}
			assert.Equal(t, tc.Request.Value, value)
			if tc.Request.MountPath != "" {
				assert.True(t, req.MountPath.IsSet())
				assert.Equal(t, tc.Request.MountPath, *req.MountPath.Get())
			}
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
		{
			TestName: "success_with_mount_path",
			Request: variable.UpsertRequest{
				Key:       gofakeit.Word(),
				Value:     gofakeit.Word(),
				MountPath: "/etc/config/app.yaml",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryEnvironmentVariableEditRequestFromDomain(tc.Request)
			value := ""
			if req.Value != nil {
				value = *req.Value
			}
			assert.Equal(t, tc.Request.Key, req.Key)
			assert.Equal(t, tc.Request.Value, value)
			if tc.Request.MountPath != "" {
				assert.True(t, req.MountPath.IsSet())
				assert.Equal(t, tc.Request.MountPath, *req.MountPath.Get())
			}
		})
	}
}
