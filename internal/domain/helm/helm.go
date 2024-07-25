package helm

import (
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

const (
	DefaultTimeoutSec int64 = 600
)

var (
	ErrInvalidHelm                          = errors.New("invalid helm")
	ErrInvalidHelmEnvironmentIDParam        = errors.New("invalid environment id param")
	ErrInvalidHelmIDParam                   = errors.New("invalid helm id param")
	ErrInvalidHelmNameParam                 = errors.New("invalid name param")
	ErrInvalidHelmStateParam                = errors.New("invalid state param")
	ErrInvalidHelmUpsertRequest             = errors.New("invalid helm upsert request")
	ErrInvalidHelmEnvironmentVariablesParam = errors.New("invalid helm environment variables param")
	ErrInvalidHelmSecretsParam              = errors.New("invalid helm secrets param")
	ErrInvalidHelmValuesOverride            = errors.New("invalid helm values override param")
	ErrInvalidPortProtocol                  = errors.New("invalid port protocol param")
	ErrInvalidUpsertRequest                 = errors.New("invalid helm upsert request")
	ErrFailedToSetHosts                     = errors.New("failed to set hosts")
)

type Helm struct {
	ID                           uuid.UUID `validate:"required"`
	EnvironmentID                uuid.UUID `validate:"required"`
	Name                         string
	IconUri                      string
	TimeoutSec                   *int32
	AutoPreview                  bool
	AutoDeploy                   bool
	Arguments                    []string
	AllowClusterWideResources    bool
	Source                       Source
	ValuesOverride               ValuesOverride
	Ports                        []Port
	BuiltInEnvironmentVariables  variable.Variables
	EnvironmentVariables         variable.Variables
	EnvironmentVariableAliases   variable.Variables
	EnvironmentVariableOverrides variable.Variables
	Secrets                      secret.Secrets
	SecretAliases                secret.Secrets
	SecretOverrides              secret.Secrets
	InternalHost                 *string
	ExternalHost                 *string
	State                        status.State
	DeploymentStageID            string
	AdvancedSettingsJson         string
	JobDeploymentRestrictions    []deploymentrestriction.ServiceDeploymentRestriction
	CustomDomains                []*qovery.CustomDomain
}

type SourceResponse struct {
	Git        *qovery.HelmSourceGitResponse
	Repository *qovery.HelmSourceRepositoryResponse
}

func (h Helm) Validate() error {
	if h.Name == "" {
		return ErrInvalidHelmNameParam
	}

	for _, ev := range h.EnvironmentVariables {
		if err := ev.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidHelmEnvironmentVariablesParam.Error())
		}
	}

	for _, sec := range h.Secrets {
		if err := sec.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidHelmSecretsParam.Error())
		}
	}

	if err := validator.New().Struct(h); err != nil {
		return errors.Wrap(err, ErrInvalidHelm.Error())
	}

	return nil
}

func (h Helm) IsValid() bool {
	return h.Validate() == nil
}

type NewHelmParams struct {
	HelmID                    string
	EnvironmentID             string
	Name                      string
	IconUri                   string
	TimeoutSec                *int32
	AutoPreview               bool
	AutoDeploy                bool
	Arguments                 []string
	AllowClusterWideResources bool
	Source                    NewHelmSourceParams
	ValuesOverride            NewHelmValuesOverrideParams
	Ports                     []NewHelmPortParams
	State                     *string
	EnvironmentVariables      variable.NewVariablesParams
	Secrets                   secret.NewSecretsParams
	DeploymentStageID         string
	AdvancedSettingsJson      string
	CustomDomains             []*qovery.CustomDomain
}

func NewHelm(params NewHelmParams) (*Helm, error) {
	helmUUID, err := uuid.Parse(params.HelmID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmIDParam.Error())
	}

	environmentUUID, err := uuid.Parse(params.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmEnvironmentIDParam.Error())
	}

	source, err := NewHelmSource(params.Source)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmValuesOverride.Error())
	}

	valuesOverride, err := NewHelmValuesOverride(params.ValuesOverride)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmValuesOverride.Error())
	}

	ports := make([]Port, 0, len(params.Ports))
	for _, p := range params.Ports {
		protocol, err := NewProtocolFromString(p.Protocol)
		if err != nil {
			return nil, errors.Wrap(err, port.ErrInvalidProtocolParam.Error())
		}

		ports = append(ports, Port{
			Name:         p.Name,
			ServiceName:  p.ServiceName,
			Namespace:    p.Namespace,
			InternalPort: p.InternalPort,
			ExternalPort: p.ExternalPort,
			Protocol:     *protocol,
			IsDefault:    p.IsDefault,
		})
	}

	h := &Helm{
		ID:                        helmUUID,
		EnvironmentID:             environmentUUID,
		Name:                      params.Name,
		IconUri:                   params.IconUri,
		TimeoutSec:                params.TimeoutSec,
		AutoPreview:               params.AutoPreview,
		AutoDeploy:                params.AutoDeploy,
		Arguments:                 params.Arguments,
		AllowClusterWideResources: params.AllowClusterWideResources,
		Source:                    *source,
		ValuesOverride:            *valuesOverride,
		Ports:                     ports,
		DeploymentStageID:         params.DeploymentStageID,
		AdvancedSettingsJson:      params.AdvancedSettingsJson,
		CustomDomains:             params.CustomDomains,
	}

	environmentVariables := make(variable.Variables, len(params.EnvironmentVariables))
	for idx, ev := range params.EnvironmentVariables {
		environmentVariable, err := variable.NewVariable(ev)
		environmentVariables[idx] = *environmentVariable
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidHelmEnvironmentVariablesParam.Error())
		}
	}
	if err := h.SetEnvironmentVariables(environmentVariables); err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmEnvironmentVariablesParam.Error())
	}

	secrets := make(secret.Secrets, len(params.Secrets))
	for idx, s := range params.Secrets {
		newSecret, err := secret.NewSecret(s)
		secrets[idx] = *newSecret
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidHelmSecretsParam.Error())
		}
	}
	if err := h.SetSecrets(secrets); err != nil {
		return nil, errors.Wrap(err, ErrInvalidHelmSecretsParam.Error())
	}

	if params.State != nil {
		helmState, err := status.NewStateFromString(*params.State)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidHelmStateParam.Error())
		}

		if err := h.SetState(*helmState); err != nil {
			return nil, errors.Wrap(err, ErrInvalidHelmStateParam.Error())
		}
	}

	if err := h.Validate(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Helm) SetEnvironmentVariables(vars variable.Variables) error {
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

	h.EnvironmentVariables = envVars
	h.BuiltInEnvironmentVariables = builtIn

	if err := h.SetHosts(vars); err != nil {
		return err
	}

	return nil
}

func (h *Helm) SetSecrets(secrets secret.Secrets) error {
	if err := secrets.Validate(); err != nil {
		return err
	}

	helmSecrets := make(secret.Secrets, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope == variable.ScopeBuiltIn {
			continue
		}
		helmSecrets = append(helmSecrets, s)
	}

	h.Secrets = helmSecrets

	return nil
}

func (h *Helm) SetState(st status.State) error {
	if err := st.Validate(); err != nil {
		return err
	}

	if st == status.StateReady {
		st = status.StateStopped
	}

	h.State = st

	return nil
}

func (h *Helm) SetHosts(vars variable.Variables) error {
	if len(vars) == 0 {
		return nil
	}

	hostExternalKey := fmt.Sprintf("QOVERY_HELM_Z%s_HOST_EXTERNAL", strings.ToUpper(strings.Split(h.ID.String(), "-")[0]))
	hostInternalKey := fmt.Sprintf("QOVERY_HELM_Z%s_HOST_INTERNAL", strings.ToUpper(strings.Split(h.ID.String(), "-")[0]))

	for _, v := range vars {
		if v.Key == hostExternalKey {
			h.ExternalHost = pointer.ToString(v.Value)
			continue
		}
		if v.Key == hostInternalKey {
			h.InternalHost = pointer.ToString(v.Value)
			continue
		}
		if h.ExternalHost != nil && h.InternalHost != nil {
			return nil
		}
	}

	// One of hte host_external or the host_internal has to be present
	if h.ExternalHost == nil && h.InternalHost == nil {
		return ErrFailedToSetHosts
	}

	return nil
}
