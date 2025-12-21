package terraformservice

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	DefaultCPU        int32  = 1000
	MinCPU            int32  = 10
	DefaultRAM        int32  = 1024
	MinRAM            int32  = 1
	DefaultGPU        int32  = 0
	MinGPU            int32  = 0
	DefaultStorage    int32  = 20
	MinStorage        int32  = 1
	DefaultRootPath   string = "/"
	DefaultIconURI    string = "app://qovery-console/terraform"
	DefaultTimeoutSec int32  = 1800
	MinTimeoutSec     int32  = 0
)

var (
	// ErrNilTerraformService is returned if a TerraformService is nil.
	ErrNilTerraformService = errors.New("terraform service cannot be nil")
	// ErrInvalidTerraformService is the error return if a TerraformService is invalid.
	ErrInvalidTerraformService = errors.New("invalid terraform service")
	// ErrInvalidEnvironmentIDParam is returned if the environment id param is invalid.
	ErrInvalidTerraformServiceEnvironmentIDParam = errors.New("invalid environment id param")
	// ErrInvalidTerraformServiceIDParam is returned if the terraform service id param is invalid.
	ErrInvalidTerraformServiceIDParam = errors.New("invalid terraform service id param")
	// ErrInvalidNameParam is returned if the name param is invalid.
	ErrInvalidTerraformServiceNameParam = errors.New("invalid name param: must contain at least one ASCII letter")
	// ErrInvalidDescriptionParam is returned if the description param is invalid.
	ErrInvalidTerraformServiceDescriptionParam = errors.New("invalid description param")
	// ErrInvalidGitRepositoryParam is returned if the git repository param is invalid.
	ErrInvalidTerraformServiceGitRepositoryParam = errors.New("invalid git repository param")
	// ErrInvalidTfVarPathParam is returned if a tfvar path is invalid.
	ErrInvalidTerraformServiceTfVarPathParam = errors.New("invalid tfvar path: must start with root_path")
	// ErrInvalidVariableParam is returned if a variable is invalid.
	ErrInvalidTerraformServiceVariableParam = errors.New("invalid variable param")
	// ErrInvalidBackendParam is returned if the backend param is invalid.
	ErrInvalidTerraformServiceBackendParam = errors.New("invalid backend configuration")
	// ErrMissingBackendType is returned if no backend type is specified.
	ErrMissingBackendType = errors.New("exactly one backend type must be specified: kubernetes or user_provided")
	// ErrMultipleBackendTypes is returned if multiple backend types are specified.
	ErrMultipleBackendTypes = errors.New("cannot specify multiple backend types")
	// ErrInvalidEngineParam is returned if the engine param is invalid.
	ErrInvalidTerraformServiceEngineParam = errors.New("invalid engine param")
	// ErrInvalidEngineVersionParam is returned if the engine version param is invalid.
	ErrInvalidTerraformServiceEngineVersionParam = errors.New("invalid engine version param")
	// ErrInvalidJobResourcesParam is returned if the job resources param is invalid.
	ErrInvalidTerraformServiceJobResourcesParam = errors.New("invalid job resources param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidTerraformServiceUpsertRequest = errors.New("invalid terraform service upsert request")
)

// Engine represents the Terraform engine type
type Engine string

const (
	EngineTerraform Engine = "TERRAFORM"
	EngineOpenTofu  Engine = "OPEN_TOFU"
)

// Validate validates the Engine
func (e Engine) Validate() error {
	switch e {
	case EngineTerraform, EngineOpenTofu:
		return nil
	default:
		return ErrInvalidTerraformServiceEngineParam
	}
}

// TerraformService represents a Terraform service domain entity
type TerraformService struct {
	ID                    uuid.UUID `validate:"required"`
	EnvironmentID         uuid.UUID `validate:"required"`
	DeploymentStageID     string
	Name                  string `validate:"required"`
	Description           *string
	AutoDeploy            bool
	GitRepository         GitRepository `validate:"required"`
	TfVarFiles            []string
	Variables             []Variable
	Backend               Backend       `validate:"required"`
	Engine                Engine        `validate:"required"`
	EngineVersion         EngineVersion `validate:"required"`
	JobResources          JobResources  `validate:"required"`
	TimeoutSec            *int32
	IconURI               string
	UseClusterCredentials bool
	ActionExtraArguments  map[string][]string
	AdvancedSettingsJson  string
	CreatedAt             time.Time
	UpdatedAt             *time.Time
}

// GitRepository represents the git repository configuration
type GitRepository struct {
	URL        string `validate:"required"`
	Branch     string
	RootPath   string
	GitTokenID *uuid.UUID
}

// Validate validates the GitRepository
func (g GitRepository) Validate() error {
	if g.URL == "" {
		return errors.New("git repository URL is required")
	}

	// Check for directory traversal
	if strings.Contains(g.RootPath, "..") || strings.Contains(g.RootPath, "~") {
		return errors.New("root_path cannot contain directory traversal sequences (.., ~)")
	}

	if err := validator.New().Struct(g); err != nil {
		return errors.Wrap(err, "invalid git repository")
	}

	return nil
}

// Variable represents a Terraform variable
type Variable struct {
	Key    string `validate:"required"`
	Value  string `validate:"required"`
	Secret bool
}

// Validate validates the Variable
func (v Variable) Validate() error {
	if v.Key == "" {
		return errors.New("variable key is required")
	}
	if v.Value == "" {
		return errors.New("variable value is required")
	}
	return nil
}

// Backend represents the Terraform backend configuration
type Backend struct {
	Kubernetes   *KubernetesBackend
	UserProvided *UserProvidedBackend
}

// KubernetesBackend represents a Kubernetes backend (empty struct)
type KubernetesBackend struct{}

// UserProvidedBackend represents a user-provided backend (empty struct)
type UserProvidedBackend struct{}

// Validate validates the Backend
func (b Backend) Validate() error {
	hasKubernetes := b.Kubernetes != nil
	hasUserProvided := b.UserProvided != nil

	if !hasKubernetes && !hasUserProvided {
		return ErrMissingBackendType
	}

	if hasKubernetes && hasUserProvided {
		return ErrMultipleBackendTypes
	}

	return nil
}

// EngineVersion represents the Terraform/OpenTofu engine version configuration
type EngineVersion struct {
	ExplicitVersion        string `validate:"required"`
	ReadFromTerraformBlock bool
}

// Validate validates the EngineVersion
func (p EngineVersion) Validate() error {
	if p.ExplicitVersion == "" {
		return errors.New("explicit_version is required")
	}

	if err := validator.New().Struct(p); err != nil {
		return errors.Wrap(err, "invalid engine version")
	}

	return nil
}

// JobResources represents the job resource limits
type JobResources struct {
	CPUMilli   int32 `validate:"required,min=10"`
	RAMMiB     int32 `validate:"required,min=1"`
	GPU        int32 `validate:"min=0"`
	StorageGiB int32 `validate:"required,min=1"`
}

// Validate validates the JobResources
func (j JobResources) Validate() error {
	if j.CPUMilli < MinCPU {
		return errors.New(fmt.Sprintf("cpu_milli must be at least %d", MinCPU))
	}
	if j.RAMMiB < MinRAM {
		return errors.New(fmt.Sprintf("ram_mib must be at least %d", MinRAM))
	}
	if j.GPU < MinGPU {
		return errors.New(fmt.Sprintf("gpu must be at least %d", MinGPU))
	}
	if j.StorageGiB < MinStorage {
		return errors.New(fmt.Sprintf("storage_gib must be at least %d", MinStorage))
	}

	if err := validator.New().Struct(j); err != nil {
		return errors.Wrap(err, "invalid job resources")
	}

	return nil
}

// Validate returns an error to tell whether the TerraformService domain model is valid or not.
func (t TerraformService) Validate() error {
	// Validate name contains at least one ASCII letter
	if t.Name == "" {
		return ErrInvalidTerraformServiceNameParam
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(t.Name)
	if !hasLetter {
		return ErrInvalidTerraformServiceNameParam
	}

	// Validate git repository
	if err := t.GitRepository.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceGitRepositoryParam.Error())
	}

	// Validate tfvar file paths
	for _, tfVarPath := range t.TfVarFiles {
		if !strings.HasPrefix(tfVarPath, t.GitRepository.RootPath) {
			return errors.Wrap(
				errors.New(fmt.Sprintf("tfvar path %q must start with root_path %q", tfVarPath, t.GitRepository.RootPath)),
				ErrInvalidTerraformServiceTfVarPathParam.Error(),
			)
		}
		// Check for directory traversal
		if strings.Contains(tfVarPath, "..") || strings.Contains(tfVarPath, "~") {
			return errors.Wrap(
				errors.New("tfvar path cannot contain directory traversal sequences (.., ~)"),
				ErrInvalidTerraformServiceTfVarPathParam.Error(),
			)
		}
	}

	// Validate variables
	for _, variable := range t.Variables {
		if err := variable.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidTerraformServiceVariableParam.Error())
		}
	}

	// Validate backend
	if err := t.Backend.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceBackendParam.Error())
	}

	// Validate engine
	if err := t.Engine.Validate(); err != nil {
		return err
	}

	// Validate engine version
	if err := t.EngineVersion.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceEngineVersionParam.Error())
	}

	// Validate job resources
	if err := t.JobResources.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformServiceJobResourcesParam.Error())
	}

	// Validate timeout
	if t.TimeoutSec != nil && *t.TimeoutSec < MinTimeoutSec {
		return errors.New(fmt.Sprintf("timeout_sec must be at least %d", MinTimeoutSec))
	}

	// Struct tag validation
	if err := validator.New().Struct(t); err != nil {
		return errors.Wrap(err, ErrInvalidTerraformService.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the TerraformService domain model is valid or not.
func (t TerraformService) IsValid() bool {
	return t.Validate() == nil
}
