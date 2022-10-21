package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainEnvironmentFromQovery(e *qovery.Environment) (*environment.Environment, error) {
	if e == nil {
		return nil, variable.ErrNilVariable
	}

	return environment.NewEnvironment(environment.NewEnvironmentParams{
		EnvironmentID: e.Id,
		ProjectID:     e.Project.Id,
		ClusterID:     e.ClusterId,
		Name:          e.Name,
		Mode:          string(e.Mode),
	})
}

// newQoveryEnvironmentVariableRequestFromDomain takes the domain request variable.UpsertRequest and turns it into a qovery.EnvironmentVariableRequest to make the api call.
func newQoveryCreateEnvironmentRequestFromDomain(request environment.CreateRepositoryRequest) (*qovery.CreateEnvironmentRequest, error) {
	mode, err := newQoveryCreateEnvironmentModeEnumFromDomain(request.Mode)
	if err != nil {
		return nil, err
	}

	return &qovery.CreateEnvironmentRequest{
		Name:    request.Name,
		Cluster: request.ClusterID,
		Mode:    mode,
	}, nil
}

// newQoveryEnvironmentEditRequestFromDomain takes the domain request environment.C and turns it into a qovery.EnvironmentVariableRequest to make the api call.
func newQoveryEnvironmentEditRequestFromDomain(request environment.UpdateRepositoryRequest) (*qovery.EnvironmentEditRequest, error) {
	mode, err := newQoveryCreateEnvironmentModeEnumFromDomain(request.Mode)
	if err != nil {
		return nil, err
	}

	return &qovery.EnvironmentEditRequest{
		Name: request.Name,
		Mode: mode,
	}, nil
}

func newQoveryCreateEnvironmentModeEnumFromDomain(mode *environment.Mode) (*qovery.CreateEnvironmentModeEnum, error) {
	if mode == nil {
		return nil, nil
	}

	m, err := qovery.NewCreateEnvironmentModeEnumFromValue(mode.String())
	if err != nil {
		return nil, errors.Wrap(err, environment.ErrInvalidModeParam.Error())
	}

	return m, nil
}
