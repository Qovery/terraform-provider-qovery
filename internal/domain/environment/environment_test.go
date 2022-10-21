package environment_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
)

func TestNewEnvironment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        environment.NewEnvironmentParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_environment_id",
			Params: environment.NewEnvironmentParams{
				ProjectID: gofakeit.UUID(),
				ClusterID: gofakeit.UUID(),
				Name:      gofakeit.Name(),
				Mode:      environment.ModeDevelopment.String(),
			},
			ExpectedError: environment.ErrInvalidEnvironmentIDParam,
		},
		{
			TestName: "fail_with_invalid_envect_id",
			Params: environment.NewEnvironmentParams{
				EnvironmentID: gofakeit.UUID(),
				ClusterID:     gofakeit.UUID(),
				Name:          gofakeit.Name(),
				Mode:          environment.ModeDevelopment.String(),
			},
			ExpectedError: environment.ErrInvalidProjectIDParam,
		},
		{
			TestName: "fail_with_invalid_cluster_id",
			Params: environment.NewEnvironmentParams{
				EnvironmentID: gofakeit.UUID(),
				ProjectID:     gofakeit.UUID(),
				Name:          gofakeit.Name(),
				Mode:          environment.ModeDevelopment.String(),
			},
			ExpectedError: environment.ErrInvalidClusterIDParam,
		},
		{
			TestName: "fail_with_invalid_name",
			Params: environment.NewEnvironmentParams{
				EnvironmentID: gofakeit.UUID(),
				ProjectID:     gofakeit.UUID(),
				ClusterID:     gofakeit.UUID(),
				Mode:          environment.ModeDevelopment.String(),
			},
			ExpectedError: environment.ErrInvalidNameParam,
		},
		{
			TestName: "fail_with_invalid_mode",
			Params: environment.NewEnvironmentParams{
				EnvironmentID: gofakeit.UUID(),
				ProjectID:     gofakeit.UUID(),
				ClusterID:     gofakeit.UUID(),
				Name:          gofakeit.Name(),
			},
			ExpectedError: environment.ErrInvalidModeParam,
		},
		{
			TestName: "success",
			Params: environment.NewEnvironmentParams{
				EnvironmentID: gofakeit.UUID(),
				ProjectID:     gofakeit.UUID(),
				ClusterID:     gofakeit.UUID(),
				Name:          gofakeit.Name(),
				Mode:          environment.ModeDevelopment.String(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			env, err := environment.NewEnvironment(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, env)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, env)
			assert.True(t, env.IsValid())
			assert.Equal(t, tc.Params.EnvironmentID, env.ID.String())
			assert.Equal(t, tc.Params.ProjectID, env.ProjectID.String())
			assert.Equal(t, tc.Params.ClusterID, env.ClusterID.String())
			assert.Equal(t, tc.Params.Name, env.Name)
			assert.Equal(t, tc.Params.Mode, env.Mode.String())
			assert.Len(t, tc.Params.EnvironmentVariables, len(env.EnvironmentVariables))
			assert.Len(t, tc.Params.Secrets, len(env.Secrets))
		})
	}
}
