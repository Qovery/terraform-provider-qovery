package newdeployment

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldFailWhenCreatingNewIncoherentDeployment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName           string
		Params             NewDeploymentParams
		ExpectedError      error
		ExpectedDeployment *Deployment
	}{
		{
			TestName: "should_fail_with_no_environment_id",
			Params: NewDeploymentParams{
				DesiredState: "RUNNING",
			},
			ExpectedError:      ErrInvalidEnvironmentIdParam,
			ExpectedDeployment: nil,
		},
		{
			TestName: "should_fail_with_wrong_desired_state",
			Params: NewDeploymentParams{
				DesiredState: "WRONG_DESIRED_STATE",
			},
			ExpectedError:      ErrInvalidDeployment,
			ExpectedDeployment: nil,
		},
		{
			TestName: "should_fail_with_wrong_environment_id",
			Params: NewDeploymentParams{
				EnvironmentId: "WRONG_UUID",
				DesiredState:  "RUNNING",
			},
			ExpectedError:      ErrInvalidEnvironmentIdParam,
			ExpectedDeployment: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			deployment, err := NewDeployment(tc.Params)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tc.ExpectedError.Error())
			assert.Nil(t, deployment)
		})
	}
}

func TestShouldCreateNewEnvironmentDeployment(t *testing.T) {
	t.Run("should_create_environment_deployment", func(t *testing.T) {
		params := NewDeploymentParams{
			EnvironmentId: uuid.NewString(),
			DesiredState:  "RUNNING",
		}

		deployment, err := NewDeployment(params)

		expectedEnvironmentId, _ := uuid.Parse(params.EnvironmentId)

		assert.NoError(t, err)
		assert.NotNil(t, deployment)
		assert.Equal(t, &expectedEnvironmentId, deployment.EnvironmentId)
		assert.Equal(t, RUNNING, deployment.DesiredState)
	})
}
