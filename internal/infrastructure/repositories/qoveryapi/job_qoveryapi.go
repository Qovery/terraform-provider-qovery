package qoveryapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

// Ensure jobQoveryAPI defined types fully satisfy the job.Repository interface.
var _ job.Repository = jobQoveryAPI{}

// jobQoveryAPI implements the interface job.Repository.
type jobQoveryAPI struct {
	client *qovery.APIClient
}

// newJobQoveryAPI return a new instance of a job.Repository that uses Qovery's API.
func newJobQoveryAPI(client *qovery.APIClient) (job.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &jobQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a job for an organization using the given organizationID and request.
func (c jobQoveryAPI) Create(ctx context.Context, environmentID string, request job.UpsertRepositoryRequest) (*job.Job, error) {
	req, err := newQoveryJobRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrInvalidJobUpsertRequest.Error())
	}

	newJob, resp, err := c.client.JobsAPI.
		CreateJob(ctx, environmentID).
		JobRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceJob, request.Name, resp, err)
	}

	var newJobId string
	if newJob.CronJobResponse != nil {
		newJobId = newJob.CronJobResponse.Id
	} else {
		newJobId = newJob.LifecycleJobResponse.Id
	}

	// The job exists in Qovery from here on. partial carries what is already known about
	// it, so that a failure in any of the calls below still reaches the Terraform state.
	// Dropping it would orphan the job and make every later apply fail with "a job named
	// X already exists".
	partial, partialErr := newDomainJobFromQovery(newJob, request.DeploymentStageID, request.IsSkipped, request.AdvancedSettingsJson)
	if partialErr != nil {
		// The response cannot be represented as a domain job, but the job does exist. Fall
		// back to its identifiers: writing the ID is the whole point here.
		partial = identityOnlyJob(newJob)
	}

	// Attach job to deployment stage
	if len(request.DeploymentStageID) > 0 {
		response, err := attachServiceToDeploymentStage(ctx, c.client, request.DeploymentStageID, newJobId, request.IsSkipped)
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return partial, apierrors.NewCreateAPIError(apierrors.APIResourceJob, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.JOB, newJobId, request.AdvancedSettingsJson)
	if err != nil {
		return partial, apierrors.NewCreateAPIError(apierrors.APIResourceJob, request.Name, nil, err)
	}

	// Get job deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, newJobId).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return partial, apierrors.NewCreateAPIError(apierrors.APIResourceJob, newJobId, resp, err)
	}

	return newDomainJobFromQovery(newJob, deploymentStage.Id, getServiceIsSkipped(deploymentStage, newJobId), request.AdvancedSettingsJson)
}

// Get calls Qovery's API to retrieve a job using the given jobID.
func (c jobQoveryAPI) Get(ctx context.Context, jobID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*job.Job, error) {
	job, resp, err := c.client.JobMainCallsAPI.
		GetJob(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	// Get job deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, jobID).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).ReadServiceAdvancedSettings(domain.JOB, jobID, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceJob, jobID, nil, err)
	}

	return newDomainJobFromQovery(job, deploymentStage.Id, getServiceIsSkipped(deploymentStage, jobID), *advancedSettingsAsJson)
}

// Update calls Qovery's API to update a job using the given jobID and request.
func (c jobQoveryAPI) Update(ctx context.Context, jobID string, request job.UpsertRepositoryRequest) (*job.Job, error) {
	req, err := newQoveryJobRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, job.ErrInvalidJobUpsertRequest.Error())
	}

	job, resp, err := c.client.JobMainCallsAPI.
		EditJob(ctx, jobID).
		JobRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	// Attach job to deployment stage
	if len(request.DeploymentStageID) > 0 {
		response, err := attachServiceToDeploymentStage(ctx, c.client, request.DeploymentStageID, jobID, request.IsSkipped)
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceJob, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.JOB, jobID, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceJob, request.Name, nil, err)
	}

	// Get job deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, jobID).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	return newDomainJobFromQovery(job, deploymentStage.Id, getServiceIsSkipped(deploymentStage, jobID), request.AdvancedSettingsJson)
}

// Delete calls Qovery's API to deletes a job using the given jobID.
func (c jobQoveryAPI) Delete(ctx context.Context, jobID string) error {
	_, resp, err := c.client.JobMainCallsAPI.
		GetJob(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// if the job is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	resp, err = c.client.JobMainCallsAPI.
		DeleteJob(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceJob, jobID, resp, err)
	}

	return nil
}

// identityOnlyJob is the last resort when a freshly created job cannot be converted from
// its API response — job.NewJob validates the schedule and source the API returned, which
// can reject a response the server itself accepted. Only the identifiers matter: the
// resource layer needs the ID in the Terraform state so the job gets tainted and replaced
// rather than orphaned. Returns nil when even the identifiers make no sense.
func identityOnlyJob(j *qovery.JobResponse) *job.Job {
	if j == nil {
		return nil
	}

	var id, environment, name string
	switch {
	case j.CronJobResponse != nil:
		id, environment, name = j.CronJobResponse.Id, j.CronJobResponse.Environment.Id, j.CronJobResponse.Name
	case j.LifecycleJobResponse != nil:
		id, environment, name = j.LifecycleJobResponse.Id, j.LifecycleJobResponse.Environment.Id, j.LifecycleJobResponse.Name
	default:
		return nil
	}

	jobID, err := uuid.Parse(id)
	if err != nil {
		return nil
	}

	environmentID, err := uuid.Parse(environment)
	if err != nil {
		return nil
	}

	return &job.Job{
		ID:            jobID,
		EnvironmentID: environmentID,
		Name:          name,
	}
}
