package project

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	// ErrNilProject is returned if a Project is nil.
	ErrNilProject = errors.New("variable cannot be nil")
	// ErrInvalidProject is the error return if a Project is invalid.
	ErrInvalidProject = errors.New("invalid project")
	// ErrInvalidOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidProjectIDParam is returned if the project id param is invalid.
	ErrInvalidProjectIDParam = errors.New("invalid project id param")
	// ErrInvalidProjectOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidProjectOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidProjectNameParam is returned if the value param is invalid.
	ErrInvalidProjectNameParam = errors.New("invalid project name param")
	// ErrInvalidProjectEnvironmentVariablesParam is returned if the environment variables param is invalid.
	ErrInvalidProjectEnvironmentVariablesParam = errors.New("invalid project environment variables param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid project upsert request")
)

type Project struct {
	ID                          uuid.UUID
	OrganizationID              uuid.UUID
	Name                        string
	Description                 *string
	EnvironmentVariables        variable.Variables
	BuiltInEnvironmentVariables variable.Variables
	Secrets                     secret.Secrets
}

// Validate returns an error to tell whether the Project domain model is valid or not.
func (p Project) Validate() error {
	return validator.New().Struct(p)
}

// IsValid returns a bool to tell whether the Project domain model is valid or not.
func (p Project) IsValid() bool {
	return p.Validate() == nil
}

// NewProjectParams represents the arguments needed to create a Project.
type NewProjectParams struct {
	ProjectID            string
	OrganizationID       string
	Name                 string
	Description          *string
	EnvironmentVariables variable.Variables
	Secrets              secret.Secrets
}

// NewProject returns a new instance of a Project domain model.
func NewProject(params NewProjectParams) (*Project, error) {
	projectUUID, err := uuid.Parse(params.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidProjectIDParam.Error())
	}

	organizationUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidProjectOrganizationIDParam.Error())
	}

	if err := params.EnvironmentVariables.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidProjectEnvironmentVariablesParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidProjectNameParam
	}

	v := &Project{
		ID:             projectUUID,
		OrganizationID: organizationUUID,
		Name:           params.Name,
		Description:    params.Description,
	}

	v.SetEnvironmentVariables(params.EnvironmentVariables)
	v.SetSecrets(params.Secrets)

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidProject.Error())
	}

	return v, nil
}

// SetEnvironmentVariables takes a variable.Variables and sets the attributes EnvironmentVariables & BuiltInEnvironmentVariables by splitting the variable with the `BUILT_IN` scope from the others.
func (p *Project) SetEnvironmentVariables(vars variable.Variables) error {
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

func (p *Project) SetSecrets(secrets secret.Secrets) error {
	if err := secrets.Validate(); err != nil {
		return err
	}

	projectSecrets := make(secret.Secrets, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope == variable.ScopeBuiltIn {
			continue
		}
		projectSecrets = append(projectSecrets, s)
	}

	p.Secrets = projectSecrets

	return nil
}
