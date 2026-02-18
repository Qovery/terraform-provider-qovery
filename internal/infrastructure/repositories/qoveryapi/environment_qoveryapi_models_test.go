package qoveryapi

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
)

func TestNewDomainEnvironmentFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Environment   *qovery.Environment
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_environment",
			Environment:   nil,
			ExpectedError: environment.ErrNilEnvironment,
		},
		{
			TestName: "success",
			Environment: &qovery.Environment{
				Id: gofakeit.UUID(),
				Project: qovery.ReferenceObject{
					Id: gofakeit.UUID(),
				},
				ClusterId: gofakeit.UUID(),
				Name:      gofakeit.Name(),
				Mode:      qovery.ENVIRONMENTMODEENUM_DEVELOPMENT,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			env, err := newDomainEnvironmentFromQovery(tc.Environment)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, env)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, env)
			assert.True(t, env.IsValid())
			assert.Equal(t, tc.Environment.Id, env.ID.String())
			assert.Equal(t, tc.Environment.ClusterId, env.ClusterID.String())
			assert.Equal(t, tc.Environment.Project.Id, env.ProjectID.String())
			assert.Equal(t, tc.Environment.Name, env.Name)
			assert.Equal(t, string(tc.Environment.Mode), env.Mode.String())
		})
	}
}

func TestNewQoveryCreateEnvironmentRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       environment.CreateRepositoryRequest
		ExpectedError error
	}{
		{
			TestName: "fail_invalid_mode",
			Request: environment.CreateRepositoryRequest{
				Name:      gofakeit.Name(),
				ClusterID: new(gofakeit.UUID()),
				Mode:      pointer.To(environment.ModePreview),
			},
			ExpectedError: environment.ErrInvalidModeParam,
		},
		{
			TestName: "success_without_cluster_id",
			Request: environment.CreateRepositoryRequest{
				Name: gofakeit.Name(),
				Mode: pointer.To(environment.ModeDevelopment),
			},
		},
		{
			TestName: "success_without_mode",
			Request: environment.CreateRepositoryRequest{
				Name:      gofakeit.Name(),
				ClusterID: new(gofakeit.UUID()),
			},
		},
		{
			TestName: "success",
			Request: environment.CreateRepositoryRequest{
				Name:      gofakeit.Name(),
				ClusterID: new(gofakeit.UUID()),
				Mode:      pointer.To(environment.ModeDevelopment),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req, err := newQoveryCreateEnvironmentRequestFromDomain(tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, req)
				return
			}

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.ClusterID, req.Cluster)
			if tc.Request.Mode == nil {
				assert.Nil(t, req.Mode)
			} else {
				assert.Equal(t, tc.Request.Mode.String(), string(*req.Mode))
			}
		})
	}
}

func TestNewQoveryEnvironmentEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       environment.UpdateRepositoryRequest
		ExpectedError error
	}{
		{
			TestName: "fail_invalid_mode",
			Request: environment.UpdateRepositoryRequest{
				Mode: pointer.To(environment.ModePreview),
			},
			ExpectedError: environment.ErrInvalidModeParam,
		},
		{
			TestName: "success_without_name",
			Request: environment.UpdateRepositoryRequest{
				Mode: pointer.To(environment.ModeDevelopment),
			},
		},
		{
			TestName: "success_without_mode",
			Request: environment.UpdateRepositoryRequest{
				Name: new(gofakeit.Name()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req, err := newQoveryEnvironmentEditRequestFromDomain(tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, req)
				return
			}

			assert.Equal(t, tc.Request.Name, req.Name)
			if tc.Request.Mode == nil {
				assert.Nil(t, req.Mode)
			} else {
				assert.Equal(t, tc.Request.Mode.String(), string(*req.Mode))
			}
		})
	}
}
