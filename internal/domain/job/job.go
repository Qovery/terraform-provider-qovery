package job

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

const (
	DefaultState           = status.StateRunning
	DefaultCPU             = 500
	MinCPU                 = 250
	DefaultMemory          = 512
	MinMemory              = 1
	MinNbRestart           = 0
	DefaultMaxNbRestart    = 3
	MinDurationSeconds     = 15
	DefaultDurationSeconds = 30
	MinStorageSize         = 1
)

var (
	// ErrNilJob is returned if a Container is nil.
	ErrNilJob = errors.New("variable cannot be nil")
	// ErrInvalidJob is the error return if a Container is invalid.
	ErrInvalidJob = errors.New("invalid job")
	// ErrInvalidEnvironmentIDParam is returned if the environment id param is invalid.
	ErrInvalidEnvironmentIDParam = errors.New("invalid environment id param")
	// ErrInvalidJobIDParam is returned if the job id param is invalid.
	ErrInvalidJobIDParam = errors.New("invalid job id param")
	// ErrInvalidNameParam is returned if the name param is invalid.
	ErrInvalidNameParam = errors.New("invalid name param")
	// ErrInvalidJobSourceParam is returned if the name param is invalid.
	ErrInvalidJobSourceParam = errors.New("invalid job source param")
	// ErrInvalidStateParam is returned if the state param is invalid.
	ErrInvalidStateParam = errors.New("invalid state param")
	// ErrInvalidSourceParam is returned if the source param is invalid.
	ErrInvalidSourceParam = errors.New("invalid source param")
	// ErrInvalidScheduleParam is returned if the schedule param is invalid.
	ErrInvalidScheduleParam = errors.New("invalid schedule param")
	// ErrInvalidPortParam is returned if the port param is invalid.
	ErrInvalidPortParam = errors.New("invalid port param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid job upsert request")
	// ErrInvalidJobEnvironmentVariablesParam is returned if the environment variables param is invalid.
	ErrInvalidJobEnvironmentVariablesParam = errors.New("invalid job environment variables param")
	// ErrInvalidJobSecretsParam is returned if the secrets param is invalid.
	ErrInvalidJobSecretsParam = errors.New("invalid job secrets param")
	// ErrFailedToSetHosts is returned if the internal & external host failed to be set.
	ErrFailedToSetHosts = errors.New("failed to set hosts")
)

type Job struct {
	ID                 uuid.UUID `validate:"required"`
	EnvironmentID      uuid.UUID `validate:"required"`
	Name               string    `validate:"required"`
	CPU                int32     `validate:"required"`
	Memory             int32     `validate:"required"`
	MaxNbRestart       uint32    `validate:"required"`
	MaxDurationSeconds uint32    `validate:"required"`
	AutoPreview        bool

	Source   JobSource   `validate:"required"`
	Schedule JobSchedule `validate:"required"`

	Ports                       port.Ports
	EnvironmentVariables        variable.Variables
	BuiltInEnvironmentVariables variable.Variables
	Secrets                     secret.Secrets
	InternalHost                *string
	ExternalHost                *string
	State                       status.State
}

// Validate returns an error to tell whether the Job domain model is valid or not.
func (j Job) Validate() error {
	if err := j.Source.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJob.Error())
	}

	if err := j.Schedule.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJob.Error())
	}

	if err := j.Ports.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJob.Error())
	}

	if err := validator.New().Struct(j); err != nil {
		return errors.Wrap(err, ErrInvalidJob.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Job domain model is valid or not.
func (j Job) IsValid() bool {
	return j.Validate() == nil
}

// NewJobParams represents the arguments needed to create a Container.
type NewJobParams struct {
	JobID              string
	EnvironmentID      string
	Name               string
	CPU                int32
	Memory             int32
	MaxNbRestart       uint32
	MaxDurationSeconds uint32
	AutoPreview        bool
	Source             NewJobSourceParams
	Schedule           NewJobScheduleParams

	State                *string
	Ports                port.NewPortsParams
	EnvironmentVariables variable.NewVariablesParams
	Secrets              secret.NewSecretsParams
}

// NewJob returns a new instance of a Job domain model.
func NewJob(params NewJobParams) (*Job, error) {
	jobUUID, err := uuid.Parse(params.JobID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobIDParam.Error())
	}

	environmentUUID, err := uuid.Parse(params.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIDParam.Error())
	}

	jobSource, err := NewJobSource(params.Source)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobSourceParam.Error())
	}

	jobSchedule, err := NewJobSchedule(params.Schedule)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobSourceParam.Error())
	}

	ports := make(port.Ports, len(params.Ports))
	for idx, p := range params.Ports {
		port, err := port.NewPort(p)
		ports[idx] = *port
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidPortParam.Error())
		}
	}

	if params.Name == "" {
		return nil, ErrInvalidNameParam
	}

	j := &Job{
		ID:                 jobUUID,
		EnvironmentID:      environmentUUID,
		Name:               params.Name,
		AutoPreview:        params.AutoPreview,
		CPU:                params.CPU,
		Memory:             params.Memory,
		MaxNbRestart:       params.MaxNbRestart,
		MaxDurationSeconds: params.MaxDurationSeconds,
		Schedule:           *jobSchedule,
		Source:             *jobSource,
		Ports:              ports,
	}

	environmentVariables := make(variable.Variables, len(params.EnvironmentVariables))
	for idx, ev := range params.EnvironmentVariables {
		environmentVariable, err := variable.NewVariable(ev)
		environmentVariables[idx] = *environmentVariable
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobEnvironmentVariablesParam.Error())
		}
	}
	if err := j.SetEnvironmentVariables(environmentVariables); err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobEnvironmentVariablesParam.Error())
	}

	secrets := make(secret.Secrets, len(params.Secrets))
	for idx, s := range params.Secrets {
		secret, err := secret.NewSecret(s)
		secrets[idx] = *secret
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobSecretsParam.Error())
		}
	}
	if err := j.SetSecrets(secrets); err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobSecretsParam.Error())
	}

	if params.State != nil {
		jobState, err := status.NewStateFromString(*params.State)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidStateParam.Error())
		}

		if err := j.SetState(*jobState); err != nil {
			return nil, errors.Wrap(err, ErrInvalidStateParam.Error())
		}
	}

	if err := j.Validate(); err != nil {
		return nil, err
	}

	return j, nil
}

// SetEnvironmentVariables takes a variable.Variables and sets the attributes EnvironmentVariables & BuiltInEnvironmentVariables by splitting the variable with the `BUILT_IN` scope from the others.
func (c *Job) SetEnvironmentVariables(vars variable.Variables) error {
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

	c.EnvironmentVariables = envVars
	c.BuiltInEnvironmentVariables = builtIn

	return nil
}

// SetSecrets takes a secret.Secrets and sets the attributes Secrets of the job.
func (c *Job) SetSecrets(secrets secret.Secrets) error {
	if err := secrets.Validate(); err != nil {
		return err
	}

	jobSecrets := make(secret.Secrets, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope == variable.ScopeBuiltIn {
			continue
		}
		jobSecrets = append(jobSecrets, s)
	}

	c.Secrets = jobSecrets

	return nil
}

// SetState takes a status.State and sets the attributes State.
func (c *Job) SetState(st status.State) error {
	if err := st.Validate(); err != nil {
		return err
	}

	if st == status.StateReady {
		st = status.StateStopped
	}

	c.State = st

	return nil
}
