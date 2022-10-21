package environment

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

const DefaultMode = ModeDevelopment

var (
	// ErrNilEnvironment is returned if an environment is nil.
	ErrNilEnvironment = errors.New("variable cannot be nil")
	// ErrInvalidEnvironment is the error return if an environment is invalid.
	ErrInvalidEnvironment = errors.New("invalid environment")
	// ErrInvalidProjectIDParam is returned if the project id param is invalid.
	ErrInvalidProjectIDParam = errors.New("invalid project id param")
	// ErrInvalidClusterIDParam is returned if the cluster id param is invalid.
	ErrInvalidClusterIDParam = errors.New("invalid cluster id param")
	// ErrInvalidEnvironmentIDParam is returned if the environment id param is invalid.
	ErrInvalidEnvironmentIDParam = errors.New("invalid environment id param")
	// ErrInvalidModeParam is returned if the mode param is invalid.
	ErrInvalidModeParam = errors.New("invalid mode param")
	// ErrInvalidNameParam is returned if the value param is invalid.
	ErrInvalidNameParam = errors.New("invalid environment name param")
	// ErrInvalidEnvironmentEnvironmentVariablesParam is returned if the environment variables param is invalid.
	ErrInvalidEnvironmentEnvironmentVariablesParam = errors.New("invalid environment environment variables param")
	// ErrInvalidEnvironmentSecretsParam is returned if the secrets param is invalid.
	ErrInvalidEnvironmentSecretsParam = errors.New("invalid environment secrets param")
	// ErrInvalidCreateRequest is returned if the create request is invalid.
	ErrInvalidCreateRequest = errors.New("invalid environment create request")
	// ErrInvalidUpdateRequest is returned if the upsert request is invalid.
	ErrInvalidUpdateRequest = errors.New("invalid environment update request")
)

type Environment struct {
	ID                          uuid.UUID
	ProjectID                   uuid.UUID
	ClusterID                   uuid.UUID
	Name                        string
	Mode                        Mode
	EnvironmentVariables        variable.Variables
	BuiltInEnvironmentVariables variable.Variables
	Secrets                     secret.Secrets
}

// Validate returns an error to tell whether the Environment domain model is valid or not.
func (p Environment) Validate() error {
	return validator.New().Struct(p)
}

// IsValid returns a bool to tell whether the Environment domain model is valid or not.
func (p Environment) IsValid() bool {
	return p.Validate() == nil
}

// NewEnvironmentParams represents the arguments needed to create an environment.
type NewEnvironmentParams struct {
	EnvironmentID        string
	ProjectID            string
	ClusterID            string
	Name                 string
	Mode                 string
	EnvironmentVariables variable.Variables
	Secrets              secret.Secrets
}

// NewEnvironment returns a new instance of an Environment domain model.
func NewEnvironment(params NewEnvironmentParams) (*Environment, error) {
	environmentUUID, err := uuid.Parse(params.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIDParam.Error())
	}

	projectUUID, err := uuid.Parse(params.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidProjectIDParam.Error())
	}

	clusterUUID, err := uuid.Parse(params.ClusterID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidClusterIDParam.Error())
	}

	mode, err := NewModeFromString(params.Mode)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidModeParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidNameParam
	}

	v := &Environment{
		ID:        environmentUUID,
		ProjectID: projectUUID,
		ClusterID: clusterUUID,
		Name:      params.Name,
		Mode:      *mode,
	}

	if err := v.SetEnvironmentVariables(params.EnvironmentVariables); err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentEnvironmentVariablesParam.Error())
	}

	if err := v.SetSecrets(params.Secrets); err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentSecretsParam.Error())
	}

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironment.Error())
	}

	return v, nil
}

// SetEnvironmentVariables takes a variable.Variables and sets the attributes EnvironmentVariables & BuiltInEnvironmentVariables by splitting the variable with the `BUILT_IN` scope from the others.
func (p *Environment) SetEnvironmentVariables(vars variable.Variables) error {
	if err := vars.Validate(); err != nil {
		return err
	}

	envVars := make(variable.Variables, 0, len(vars))
	builtIn := make(variable.Variables, 0, len(vars))

	for _, v := range vars {
		if v.Scope == variable.ScopeBuiltIn {
			builtIn = append(builtIn, v)
			continue
		}
		envVars = append(envVars, v)
	}

	p.EnvironmentVariables = envVars
	p.BuiltInEnvironmentVariables = builtIn

	return nil
}

func (p *Environment) SetSecrets(secrets secret.Secrets) error {
	if err := secrets.Validate(); err != nil {
		return err
	}

	environmentSecrets := make(secret.Secrets, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope == variable.ScopeBuiltIn {
			continue
		}
		environmentSecrets = append(environmentSecrets, s)
	}

	p.Secrets = environmentSecrets

	return nil
}
