package container

import (
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

const (
	DefaultState               = status.StateRunning
	DefaultCPU                 = 500
	MinCPU                     = 250
	DefaultMemory              = 512
	MinMemory                  = 1
	DefaultMinRunningInstances = 1
	MinMinRunningInstances     = 1
	DefaultMaxRunningInstances = 1
	MinMaxRunningInstances     = -1
	MinStorageSize             = 1
)

var (
	// ErrNilContainer is returned if a Container is nil.
	ErrNilContainer = errors.New("variable cannot be nil")
	// ErrInvalidContainer is the error return if a Container is invalid.
	ErrInvalidContainer = errors.New("invalid container")
	// ErrInvalidEnvironmentIDParam is returned if the environment id param is invalid.
	ErrInvalidEnvironmentIDParam = errors.New("invalid environment id param")
	// ErrInvalidContainerIDParam is returned if the container id param is invalid.
	ErrInvalidContainerIDParam = errors.New("invalid container id param")
	// ErrInvalidRegistryIDParam is returned if the registry id param is invalid.
	ErrInvalidRegistryIDParam = errors.New("invalid registry id param")
	// ErrInvalidImageNameParam is returned if the container image name param is invalid.
	ErrInvalidImageNameParam = errors.New("invalid image name param")
	// ErrInvalidTagParam is returned if the container tag param is invalid.
	ErrInvalidTagParam = errors.New("invalid tag param")
	// ErrInvalidNameParam is returned if the name param is invalid.
	ErrInvalidNameParam = errors.New("invalid name param")
	// ErrInvalidStateParam is returned if the state param is invalid.
	ErrInvalidStateParam = errors.New("invalid state param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid container upsert request")
	// ErrInvalidContainerEnvironmentVariablesParam is returned if the environment variables param is invalid.
	ErrInvalidContainerEnvironmentVariablesParam = errors.New("invalid container environment variables param")
	// ErrInvalidContainerSecretsParam is returned if the secrets param is invalid.
	ErrInvalidContainerSecretsParam = errors.New("invalid container secrets param")
	// ErrFailedToSetHosts is returned if the internal & external host failed to be set.
	ErrFailedToSetHosts = errors.New("failed to set hosts")
)

type Container struct {
	ID                  uuid.UUID `validate:"required"`
	EnvironmentID       uuid.UUID `validate:"required"`
	RegistryID          uuid.UUID `validate:"required"`
	Name                string    `validate:"required"`
	ImageName           string    `validate:"required"`
	Tag                 string    `validate:"required"`
	CPU                 int32     `validate:"required"`
	Memory              int32     `validate:"required"`
	MinRunningInstances int32     `validate:"required"`
	MaxRunningInstances int32     `validate:"required"`
	AutoPreview         bool

	Entrypoint                  *string
	Arguments                   []string
	Storages                    storage.Storages
	Ports                       port.Ports
	EnvironmentVariables        variable.Variables
	BuiltInEnvironmentVariables variable.Variables
	Secrets                     secret.Secrets
	InternalHost                *string
	ExternalHost                *string
	State                       status.State
	DeploymentStageId           string
}

// Validate returns an error to tell whether the Container domain model is valid or not.
func (c Container) Validate() error {
	if err := c.Storages.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidContainer.Error())
	}

	if err := c.Ports.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidContainer.Error())
	}

	if err := validator.New().Struct(c); err != nil {
		return errors.Wrap(err, ErrInvalidContainer.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Container domain model is valid or not.
func (c Container) IsValid() bool {
	return c.Validate() == nil
}

// NewContainerParams represents the arguments needed to create a Container.
type NewContainerParams struct {
	ContainerID         string
	EnvironmentID       string
	RegistryID          string
	Name                string
	ImageName           string
	Tag                 string
	CPU                 int32
	Memory              int32
	MinRunningInstances int32
	MaxRunningInstances int32
	AutoPreview         bool

	State                *string
	Entrypoint           *string
	Arguments            []string
	Storages             storage.Storages
	Ports                port.Ports
	EnvironmentVariables variable.Variables
	Secrets              secret.Secrets
	DeploymentStageId    string
}

// NewContainer returns a new instance of a Container domain model.
func NewContainer(params NewContainerParams) (*Container, error) {
	containerUUID, err := uuid.Parse(params.ContainerID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidContainerIDParam.Error())
	}

	environmentUUID, err := uuid.Parse(params.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidEnvironmentIDParam.Error())
	}

	registryUUID, err := uuid.Parse(params.RegistryID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRegistryIDParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidNameParam
	}

	if params.ImageName == "" {
		return nil, ErrInvalidImageNameParam
	}

	if params.Tag == "" {
		return nil, ErrInvalidTagParam
	}

	c := &Container{
		ID:                  containerUUID,
		EnvironmentID:       environmentUUID,
		RegistryID:          registryUUID,
		Name:                params.Name,
		ImageName:           params.ImageName,
		Tag:                 params.Tag,
		AutoPreview:         params.AutoPreview,
		Entrypoint:          params.Entrypoint,
		CPU:                 params.CPU,
		Memory:              params.Memory,
		MinRunningInstances: params.MinRunningInstances,
		MaxRunningInstances: params.MaxRunningInstances,
		Arguments:           params.Arguments,
		Storages:            params.Storages,
		Ports:               params.Ports,
		DeploymentStageId:   params.DeploymentStageId,
	}

	if err := c.SetEnvironmentVariables(params.EnvironmentVariables); err != nil {
		return nil, errors.Wrap(err, ErrInvalidContainerEnvironmentVariablesParam.Error())
	}

	if err := c.SetSecrets(params.Secrets); err != nil {
		return nil, errors.Wrap(err, ErrInvalidContainerSecretsParam.Error())
	}

	if params.State != nil {
		containerState, err := status.NewStateFromString(*params.State)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidStateParam.Error())
		}

		if err := c.SetState(*containerState); err != nil {
			return nil, errors.Wrap(err, ErrInvalidStateParam.Error())
		}
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// SetEnvironmentVariables takes a variable.Variables and sets the attributes EnvironmentVariables & BuiltInEnvironmentVariables by splitting the variable with the `BUILT_IN` scope from the others.
func (c *Container) SetEnvironmentVariables(vars variable.Variables) error {
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

	if err := c.SetHosts(vars); err != nil {
		return nil
	}

	return nil
}

// SetSecrets takes a secret.Secrets and sets the attributes Secrets of the container.
func (c *Container) SetSecrets(secrets secret.Secrets) error {
	if err := secrets.Validate(); err != nil {
		return err
	}

	containerSecrets := make(secret.Secrets, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope == variable.ScopeBuiltIn {
			continue
		}
		containerSecrets = append(containerSecrets, s)
	}

	c.Secrets = containerSecrets

	return nil
}

// SetHosts takes a variable.Variables and sets the attributes EnvironmentVariables & BuiltInEnvironmentVariables by splitting the variable with the `BUILT_IN` scope from the others.
func (c *Container) SetHosts(vars variable.Variables) error {
	hostExternalKey := fmt.Sprintf("QOVERY_CONTAINER_Z%s_HOST_EXTERNAL", strings.ToUpper(strings.Split(c.ID.String(), "-")[0]))
	hostInternalKey := fmt.Sprintf("QOVERY_CONTAINER_Z%s_HOST_INTERNAL", strings.ToUpper(strings.Split(c.ID.String(), "-")[0]))

	for _, v := range vars {
		if v.Key == hostExternalKey {
			c.ExternalHost = pointer.ToString(v.Value)
			continue
		}
		if v.Key == hostInternalKey {
			c.InternalHost = pointer.ToString(v.Value)
			continue
		}
		if c.ExternalHost != nil && c.InternalHost != nil {
			return nil
		}
	}

	return ErrFailedToSetHosts
}

// SetState takes a status.State and sets the attributes State.
func (c *Container) SetState(st status.State) error {
	if err := st.Validate(); err != nil {
		return err
	}

	if st == status.StateReady {
		st = status.StateStopped
	}

	c.State = st

	return nil
}
