package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// newQoveryTerraformRequestFromDomain converts a domain UpsertRepositoryRequest to a Qovery API TerraformRequest.
func newQoveryTerraformRequestFromDomain(request terraformservice.UpsertRepositoryRequest) (*qovery.TerraformRequest, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	// Build git repository
	gitRepo := qovery.NewTerraformGitRepositoryRequest(request.GitRepository.URL)
	if request.GitRepository.Branch != "" {
		gitRepo.Branch = &request.GitRepository.Branch
	}
	if request.GitRepository.RootPath != "" {
		gitRepo.RootPath = &request.GitRepository.RootPath
	}
	if request.GitRepository.GitTokenID != nil {
		tokenID := request.GitRepository.GitTokenID.String()
		gitRepo.GitTokenId = &tokenID
	}

	// Build terraform_files_source with nested git_repository
	filesSourceOneOf := qovery.NewTerraformRequestTerraformFilesSourceOneOf()
	filesSourceOneOf.GitRepository = gitRepo
	filesSource := qovery.TerraformRequestTerraformFilesSourceOneOfAsTerraformRequestTerraformFilesSource(filesSourceOneOf)

	// Build variables
	tfVars := make([]qovery.TerraformVarKeyValue, 0, len(request.Variables))
	for _, v := range request.Variables {
		tfVar := qovery.NewTerraformVarKeyValue()
		tfVar.Key = &v.Key
		tfVar.Value = &v.Value
		tfVar.Secret = &v.Secret
		tfVars = append(tfVars, *tfVar)
	}

	// Build terraform_variables_source
	variablesSource := qovery.NewTerraformVariablesSourceRequest(request.TfVarFiles, tfVars)

	// Build backend (oneOf)
	var backend qovery.TerraformBackend
	if request.Backend.Kubernetes != nil {
		// Kubernetes backend (empty map)
		kubernetesBackend := qovery.NewTerraformBackendOneOf(make(map[string]any))
		backend = qovery.TerraformBackendOneOfAsTerraformBackend(kubernetesBackend)
	} else if request.Backend.UserProvided != nil {
		// User-provided backend (empty map)
		userProvidedBackend := qovery.NewTerraformBackendOneOf1(make(map[string]any))
		backend = qovery.TerraformBackendOneOf1AsTerraformBackend(userProvidedBackend)
	}

	// Build engine_version
	engineVersion := qovery.NewTerraformProviderVersion(request.EngineVersion.ExplicitVersion)
	if request.EngineVersion.ReadFromTerraformBlock {
		engineVersion.ReadFromTerraformBlock = &request.EngineVersion.ReadFromTerraformBlock
	}

	// Build job_resources
	jobResources := qovery.NewTerraformRequestJobResources(
		request.JobResources.CPUMilli,
		request.JobResources.RAMMiB,
		request.JobResources.GPU,
		request.JobResources.StorageGiB,
	)

	// Build engine
	engine := qovery.TerraformEngineEnum(request.Engine)

	// Build description (handle nil pointer)
	description := ""
	if request.Description != nil {
		description = *request.Description
	}

	// Build the main request
	req := qovery.NewTerraformRequest(
		request.Name,
		description,
		request.AutoDeploy,
		filesSource,
		*variablesSource,
		backend,
		engine,
		*engineVersion,
		*jobResources,
	)

	// Optional fields
	if request.TimeoutSec != nil {
		req.TimeoutSec = request.TimeoutSec
	}

	if request.IconURI != "" {
		req.IconUri = &request.IconURI
	}

	if request.UseClusterCredentials {
		req.UseClusterCredentials = &request.UseClusterCredentials
	}

	// Action extra arguments
	if len(request.ActionExtraArguments) > 0 {
		req.ActionExtraArguments = &request.ActionExtraArguments
	}

	return req, nil
}

// newDomainTerraformServiceFromQovery converts a Qovery API TerraformResponse to a domain TerraformService.
func newDomainTerraformServiceFromQovery(response *qovery.TerraformResponse, deploymentStageID string, isSkipped bool, advancedSettingsJson string) (*terraformservice.TerraformService, error) {
	if response == nil {
		return nil, errors.New("terraform response cannot be nil")
	}

	id, err := uuid.Parse(response.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse terraform service id")
	}

	envID, err := uuid.Parse(response.Environment.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse environment id")
	}

	// Extract git repository from terraform_files_source
	// The TerraformFilesSource in response is map[string]interface{}, need to extract git_repository
	gitRepo := terraformservice.GitRepository{
		URL:      "",
		Branch:   "",
		RootPath: terraformservice.DefaultRootPath,
	}

	// Parse terraform_files_source to extract git_repository
	if response.TerraformFilesSource != nil && response.TerraformFilesSource.TerraformFilesSource != nil {
		if response.TerraformFilesSource.TerraformFilesSource.Git != nil {
			if response.TerraformFilesSource.TerraformFilesSource.Git.GitRepository != nil {
				gitRepoResp := response.TerraformFilesSource.TerraformFilesSource.Git.GitRepository
				gitRepo.URL = gitRepoResp.Url
				if gitRepoResp.Branch != nil && *gitRepoResp.Branch != "" {
					gitRepo.Branch = *gitRepoResp.Branch
				}
				if gitRepoResp.RootPath != nil && *gitRepoResp.RootPath != "" {
					gitRepo.RootPath = *gitRepoResp.RootPath
				}
				if gitRepoResp.GitTokenId.IsSet() {
					if tokenID := gitRepoResp.GitTokenId.Get(); tokenID != nil && *tokenID != "" {
						tokenUUID, err := uuid.Parse(*tokenID)
						if err == nil {
							gitRepo.GitTokenID = &tokenUUID
						}
					}
				}
			}
		}
	}

	// Extract variables from terraform_variables_source
	variables := make([]terraformservice.Variable, 0)
	if len(response.TerraformVariablesSource.TfVars) > 0 {
		for _, v := range response.TerraformVariablesSource.TfVars {
			variable := terraformservice.Variable{
				Key:    "",
				Value:  "",
				Secret: false,
			}
			if v.Key != nil {
				variable.Key = *v.Key
			}
			if v.Value != nil {
				variable.Value = *v.Value
			}
			if v.Secret != nil {
				variable.Secret = *v.Secret
			}
			variables = append(variables, variable)
		}
	}

	// Extract backend
	backend := terraformservice.Backend{}
	if response.Backend.TerraformBackendOneOf != nil {
		backend.Kubernetes = &terraformservice.KubernetesBackend{}
	} else if response.Backend.TerraformBackendOneOf1 != nil {
		backend.UserProvided = &terraformservice.UserProvidedBackend{}
	}

	// Extract engine version
	engineVersion := terraformservice.EngineVersion{
		ExplicitVersion:        response.ProviderVersion.ExplicitVersion,
		ReadFromTerraformBlock: false,
	}
	if response.ProviderVersion.ReadFromTerraformBlock != nil {
		engineVersion.ReadFromTerraformBlock = *response.ProviderVersion.ReadFromTerraformBlock
	}

	// Extract job resources
	jobResources := terraformservice.JobResources{
		CPUMilli:   response.JobResources.CpuMilli,
		RAMMiB:     response.JobResources.RamMib,
		GPU:        response.JobResources.Gpu,
		StorageGiB: response.JobResources.StorageGib,
	}

	// Extract engine
	engine := terraformservice.Engine(response.Engine)

	// Extract tfvar files
	tfVarFiles := make([]string, 0)
	if response.TerraformVariablesSource.TfVarFilePaths != nil {
		tfVarFiles = response.TerraformVariablesSource.TfVarFilePaths
	}

	// Build the domain model
	service := &terraformservice.TerraformService{
		ID:                    id,
		EnvironmentID:         envID,
		DeploymentStageID:     deploymentStageID,
		IsSkipped:             isSkipped,
		Name:                  response.Name,
		Description:           response.Description,
		AutoDeploy:            response.AutoDeploy,
		GitRepository:         gitRepo,
		TfVarFiles:            tfVarFiles,
		Variables:             variables,
		Backend:               backend,
		Engine:                engine,
		EngineVersion:         engineVersion,
		JobResources:          jobResources,
		IconURI:               response.IconUri,
		UseClusterCredentials: false,
		ActionExtraArguments:  make(map[string][]string),
		AdvancedSettingsJson:  advancedSettingsJson,
		CreatedAt:             response.CreatedAt,
	}

	// Optional fields
	if response.TimeoutSec > 0 {
		timeout := response.TimeoutSec
		service.TimeoutSec = &timeout
	}

	service.UseClusterCredentials = response.UseClusterCredentials

	if len(response.ActionExtraArguments) > 0 {
		service.ActionExtraArguments = response.ActionExtraArguments
	}

	if response.UpdatedAt != nil {
		service.UpdatedAt = response.UpdatedAt
	}

	return service, nil
}
