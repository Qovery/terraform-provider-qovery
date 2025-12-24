//go:build unit && !integration
// +build unit,!integration

package deploymentstage_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

func TestNewDeploymentStage(t *testing.T) {
	t.Parallel()

	validDeploymentStageID := uuid.NewString()
	validEnvironmentID := uuid.NewString()
	validIsAfterID := uuid.NewString()
	validIsBeforeID := uuid.NewString()

	tests := []struct {
		name          string
		params        deploymentstage.NewDeploymentStageParams
		expectError   bool
		expectedError error
	}{
		{
			name: "success with valid params",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
			},
			expectError: false,
		},
		{
			name: "success with description",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				Description:       "Test deployment stage description",
			},
			expectError: false,
		},
		{
			name: "success with is_after",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				IsAfter:           &validIsAfterID,
			},
			expectError: false,
		},
		{
			name: "success with is_before",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				IsBefore:          &validIsBeforeID,
			},
			expectError: false,
		},
		{
			name: "success with is_after and is_before",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				IsAfter:           &validIsAfterID,
				IsBefore:          &validIsBeforeID,
			},
			expectError: false,
		},
		{
			name: "fail with invalid deployment stage id",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: "invalid-uuid",
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidDeploymentStageIDParam,
		},
		{
			name: "fail with empty deployment stage id",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: "",
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidDeploymentStageIDParam,
		},
		{
			name: "fail with invalid environment id",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     "invalid-uuid",
				Name:              "test-stage",
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidEnvironmentIDParam,
		},
		{
			name: "fail with empty environment id",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     "",
				Name:              "test-stage",
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidEnvironmentIDParam,
		},
		{
			name: "fail with empty name",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "",
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidDeploymentStageNameParam,
		},
		{
			name: "fail with invalid is_after uuid",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				IsAfter:           stringPtr("invalid-uuid"),
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidIsAfterParam,
		},
		{
			name: "fail with invalid is_before uuid",
			params: deploymentstage.NewDeploymentStageParams{
				DeploymentStageID: validDeploymentStageID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-stage",
				IsBefore:          stringPtr("invalid-uuid"),
			},
			expectError:   true,
			expectedError: deploymentstage.ErrInvalidIsBeforeParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := deploymentstage.NewDeploymentStage(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedError != nil {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.params.Name, result.Name)
				assert.Equal(t, tt.params.Description, result.Description)
			}
		})
	}
}

func TestDeploymentStage_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		stage       deploymentstage.DeploymentStage
		expectError bool
	}{
		{
			name: "valid deployment stage",
			stage: deploymentstage.DeploymentStage{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-stage",
				Description:   "Test description",
			},
			expectError: false,
		},
		{
			name: "valid deployment stage with is_after",
			stage: deploymentstage.DeploymentStage{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-stage",
				IsAfter:       uuidPtr(uuid.New()),
			},
			expectError: false,
		},
		{
			name: "valid deployment stage with is_before",
			stage: deploymentstage.DeploymentStage{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-stage",
				IsBefore:      uuidPtr(uuid.New()),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stage.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create UUID pointers
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
