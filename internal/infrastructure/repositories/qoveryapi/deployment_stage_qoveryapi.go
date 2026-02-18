package qoveryapi

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

type deploymentStageQoveryAPI struct {
	client *qovery.APIClient
}

func newDeploymentStageQoveryAPI(client *qovery.APIClient) (deploymentstage.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &deploymentStageQoveryAPI{
		client: client,
	}, nil
}

func (c deploymentStageQoveryAPI) Create(ctx context.Context, environmentID string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStageCreated, resp, err := c.client.DeploymentStageMainCallsAPI.
		CreateEnvironmentDeploymentStage(ctx, environmentID).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
	}

	if request.IsAfter != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveAfterDeploymentStage(ctx, deploymentStageCreated.Id, *request.IsAfter).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	if request.IsBefore != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveBeforeDeploymentStage(ctx, deploymentStageCreated.Id, *request.IsBefore).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStageCreated.Id,
		EnvironmentID:     deploymentStageCreated.Environment.Id,
		Name:              *deploymentStageCreated.Name,
		Description:       *deploymentStageCreated.Description,
		IsAfter:           request.IsAfter,
		IsBefore:          request.IsBefore,
	})
}

func (c deploymentStageQoveryAPI) Get(ctx context.Context, environmentID string, deploymentStageID string) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetDeploymentStage(ctx, deploymentStageID).Execute()
	if deploymentStage == nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
	})
}

func (c deploymentStageQoveryAPI) GetAllByEnvironmentID(ctx context.Context, environmentID string) (*[]deploymentstage.DeploymentStage, error) {
	result, resp, err := c.client.DeploymentStageMainCallsAPI.ListEnvironmentDeploymentStage(ctx, environmentID).Execute()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 200 {
		return nil, errors.New("Wrong environment id")
	}

	var array []deploymentstage.DeploymentStage
	for _, deploymentStage := range result.Results {
		stage, err := deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
			DeploymentStageID: deploymentStage.Id,
			EnvironmentID:     deploymentStage.Environment.Id,
			Name:              *deploymentStage.Name,
			Description:       *deploymentStage.Description,
		})
		if err == nil {
			array = append(array, *stage)
		}
	}

	return &array, nil
}

func (c deploymentStageQoveryAPI) Update(ctx context.Context, deploymentStageID string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.
		EditDeploymentStage(ctx, deploymentStageID).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	if request.IsAfter != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveAfterDeploymentStage(ctx, deploymentStageID, *request.IsAfter).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	if request.IsBefore != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveBeforeDeploymentStage(ctx, deploymentStageID, *request.IsBefore).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
		IsAfter:           request.IsAfter,
		IsBefore:          request.IsBefore,
	})
}

func (c deploymentStageQoveryAPI) Delete(ctx context.Context, deploymentStageID string) error {
	// 1. Get deployment stage to retrieve environment ID
	stage, resp, err := c.client.DeploymentStageMainCallsAPI.GetDeploymentStage(ctx, deploymentStageID).Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// Stage already deleted
			return nil
		}
		return apierrors.NewReadAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	// 2. Wait for environment to be in a stable state
	if err := c.waitForEnvironmentFinalState(ctx, stage.Environment.Id); err != nil {
		return errors.Wrap(err, "failed to wait for environment final state")
	}

	// 3. Attempt deletion with exponential backoff retry
	maxRetries := 10
	backoff := 2 * time.Second

	for attempt := range maxRetries {
		resp, err = c.client.DeploymentStageMainCallsAPI.DeleteDeploymentStage(ctx, deploymentStageID).Execute()

		// Success case
		if err == nil && resp.StatusCode < 300 {
			// 4. Wait for deployment stage to be fully deleted
			if err := c.waitForDeploymentStageDeletion(ctx, deploymentStageID); err != nil {
				return errors.Wrap(err, "deployment stage delete initiated but failed to confirm deletion")
			}
			return nil
		}

		// Check if it's the "must be empty" error
		if resp != nil && resp.StatusCode == 400 {
			// Parse error to check if it's the "must be empty" error
			if err != nil && strings.Contains(err.Error(), "must empty of service") {
				// Retry with exponential backoff
				if attempt < maxRetries-1 {
					select {
					case <-time.After(backoff):
						backoff *= 2
						if backoff > 30*time.Second {
							backoff = 30 * time.Second
						}
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}

		// Other error - return immediately
		return apierrors.NewDeleteAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	// All retries exhausted
	return apierrors.NewDeleteAPIError(
		apierrors.APIResourceDeploymentStage,
		deploymentStageID,
		resp,
		errors.New("deployment stage still has services attached after retries"),
	)
}

// waitForEnvironmentFinalState polls until the environment reaches a stable state
func (c deploymentStageQoveryAPI) waitForEnvironmentFinalState(ctx context.Context, environmentID string) error {
	timeout := time.After(2 * time.Hour)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil // Timeout - proceed anyway
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, resp, err := c.client.EnvironmentMainCallsAPI.GetEnvironmentStatus(ctx, environmentID).Execute()
			if err != nil || resp.StatusCode >= 400 {
				// If we can't get status, continue anyway
				return nil
			}

			if c.isEnvironmentInFinalState(status.State) {
				return nil
			}
		}
	}
}

// isEnvironmentInFinalState checks if environment is in a stable state
func (c deploymentStageQoveryAPI) isEnvironmentInFinalState(state qovery.StateEnum) bool {
	stateStr := string(state)
	// Not in processing/waiting/queued state
	return !strings.HasSuffix(stateStr, "ING") &&
		!strings.Contains(stateStr, "_WAITING") &&
		!strings.Contains(stateStr, "_QUEUED")
}

// waitForDeploymentStageDeletion polls until the deployment stage is deleted (404)
func (c deploymentStageQoveryAPI) waitForDeploymentStageDeletion(ctx context.Context, deploymentStageID string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errors.New("timeout waiting for deployment stage deletion")
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, resp, err := c.client.DeploymentStageMainCallsAPI.GetDeploymentStage(ctx, deploymentStageID).Execute()
			if resp != nil && resp.StatusCode == 404 {
				// Stage is deleted
				return nil
			}
			if err != nil && resp != nil && resp.StatusCode >= 500 {
				// Server error - continue polling
				continue
			}
		}
	}
}
