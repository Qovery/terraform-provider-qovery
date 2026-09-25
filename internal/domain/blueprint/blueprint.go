package blueprint

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidUpsertRequest        = errors.New("invalid blueprint upsert request")
	ErrInvalidEnvironmentIDParam   = errors.New("invalid environment id param")
	ErrInvalidBlueprintIDParam     = errors.New("invalid blueprint id param")
	ErrBlankVariableName           = errors.New("blueprint variable name cannot be blank")
	ErrVariableDeclaredTwice       = errors.New("blueprint variable declared in both variables and secret_variables")
	ErrDispatchFailed              = errors.New("blueprint dispatch failed")
	ErrServiceDeploymentFailed     = errors.New("blueprint service deployment failed")
	ErrUnexpectedDispatchStatus    = errors.New("unexpected blueprint dispatch status")
	ErrUnknownServiceType          = errors.New("unknown blueprint service type")
	ErrDispatchDidNotCreateService = errors.New("blueprint dispatch succeeded without a linked service")
	ErrInvalidCatalogVersion       = errors.New("invalid blueprint catalog version: expected <provider>/<service_family>/<service_version>, e.g. AWS/postgres/17")
	ErrInvalidTag                  = errors.New("invalid blueprint tag: expected <provider>/<service_family>/<service_version>/<release>")
	ErrCatalogVersionNotFound      = errors.New("blueprint catalog version not found")
	ErrWaitTimeout                 = errors.New("blueprint wait timed out")
)

// ServiceType is the kind of service a blueprint dispatch materializes.
type ServiceType string

const (
	ServiceTypeTerraform ServiceType = "TERRAFORM"
	ServiceTypeHelm      ServiceType = "HELM"
)

// DispatchStatus is the status of an engine dispatch, a subset of the deployment statuses.
type DispatchStatus string

const (
	DispatchStatusDeploying      DispatchStatus = "DEPLOYING"
	DispatchStatusWaitingRunning DispatchStatus = "WAITING_RUNNING"
	DispatchStatusRunning        DispatchStatus = "RUNNING"
	DispatchStatusFailed         DispatchStatus = "FAILED"
	DispatchStatusInternalError  DispatchStatus = "INTERNAL_ERROR"
	DispatchStatusCanceling      DispatchStatus = "CANCELING"
	DispatchStatusCanceled       DispatchStatus = "CANCELED"
)

// IsPending includes CANCELING: the dispatch has not settled until it reaches CANCELED.
func (s DispatchStatus) IsPending() bool {
	return s == DispatchStatusDeploying || s == DispatchStatusWaitingRunning || s == DispatchStatusCanceling
}

func (s DispatchStatus) IsSuccess() bool {
	return s == DispatchStatusRunning
}

func (s DispatchStatus) IsFailure() bool {
	switch s {
	case DispatchStatusFailed, DispatchStatusInternalError, DispatchStatusCanceled:
		return true
	case DispatchStatusDeploying, DispatchStatusWaitingRunning, DispatchStatusCanceling, DispatchStatusRunning:
		return false
	}
	return false
}

type Blueprint struct {
	ID               uuid.UUID
	EnvironmentID    uuid.UUID
	Name             string
	Tag              string
	CatalogURL       string
	ServiceType      ServiceType
	ServiceID        *string
	LatestDeployment *Dispatch
	Variables        []Variable
	ServiceStatus    *ServiceStatus
}

// LastApplyFailed reports whether the persisted settings may not be what runs.
func (b Blueprint) LastApplyFailed() bool {
	// A dispatch being canceled will not apply the saved settings, even though it has not settled yet
	if b.LatestDeployment != nil && (b.LatestDeployment.Status.IsFailure() || b.LatestDeployment.Status == DispatchStatusCanceling) {
		return true
	}
	return b.ServiceStatus != nil && strings.HasSuffix(b.ServiceStatus.State, "_ERROR")
}

// CatalogVersion is a blueprint major version in the catalog, e.g. AWS/postgres/17.
type CatalogVersion struct {
	Provider       string
	ServiceFamily  string
	ServiceVersion string
}

func (v CatalogVersion) String() string {
	return v.Provider + "/" + v.ServiceFamily + "/" + v.ServiceVersion
}

// SameService ignores the major version: a major bump upgrades in place.
func (v CatalogVersion) SameService(other CatalogVersion) bool {
	return strings.EqualFold(v.Provider, other.Provider) && strings.EqualFold(v.ServiceFamily, other.ServiceFamily)
}

func (v CatalogVersion) Equal(other CatalogVersion) bool {
	return v.SameService(other) && strings.EqualFold(v.ServiceVersion, other.ServiceVersion)
}

func ParseCatalogVersion(value string) (CatalogVersion, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return CatalogVersion{}, errors.Wrapf(ErrInvalidCatalogVersion, "%q", value)
	}
	return CatalogVersion{Provider: parts[0], ServiceFamily: parts[1], ServiceVersion: parts[2]}, nil
}

// CatalogVersionFromTag drops the release from a tag like AWS/postgres/17/4.1.0.
func CatalogVersionFromTag(tag string) (CatalogVersion, error) {
	parts := strings.Split(strings.TrimSpace(tag), "/")
	if len(parts) != 4 || parts[3] == "" {
		return CatalogVersion{}, errors.Wrapf(ErrInvalidTag, "%q", tag)
	}
	return ParseCatalogVersion(strings.Join(parts[:3], "/"))
}

type CatalogEntry struct {
	CatalogVersion
	LatestTag string
}

// Dispatch is one engine run that creates or converges the blueprint's service.
type Dispatch struct {
	ID           string
	Status       DispatchStatus
	StartedAt    time.Time
	ErrorMessage *string
}

// Variable is a blueprint variable as read back from the API. Value is nil for secrets.
type Variable struct {
	Name     string
	Value    *string
	IsSecret bool
}

// SpecOverrides overrides the engine settings of the blueprint manifest.
type SpecOverrides struct {
	EngineVersion *string
	Credentials   *string
	Backend       *string
	Timeout       *int32
	CPU           *string
	RAM           *string
	Storage       *string
}

// CreationResult identifies a created blueprint and the dispatch its creation started.
type CreationResult struct {
	BlueprintID string
	DispatchID  string
}

// ServiceStatus is the deployment status of the service a blueprint materialized.
type ServiceStatus struct {
	State              string
	LastDeploymentDate *time.Time
}

type UpsertRequest struct {
	Name            string `validate:"required"`
	Tag             string `validate:"required"`
	IconURI         string `validate:"required"`
	Variables       map[string]string
	SecretVariables map[string]string
	SpecOverrides   *SpecOverrides
}

func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}
	for name := range r.Variables {
		if strings.TrimSpace(name) == "" {
			return ErrBlankVariableName
		}
		if _, ok := r.SecretVariables[name]; ok {
			return errors.Wrap(ErrVariableDeclaredTwice, name)
		}
	}
	for name := range r.SecretVariables {
		if strings.TrimSpace(name) == "" {
			return ErrBlankVariableName
		}
	}
	return nil
}

type CreateRequest struct {
	UpsertRequest
	Deploy bool
}

// UpdateRequest carries the previously applied variable names and overrides, so the
// repository can build a merge-patch that removes what the plan no longer declares.
type UpdateRequest struct {
	UpsertRequest
	PreviousVariableNames []string
	PreviousSpecOverrides *SpecOverrides
}
