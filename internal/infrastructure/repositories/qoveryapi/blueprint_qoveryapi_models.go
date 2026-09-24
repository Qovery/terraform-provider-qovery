package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

func newQoveryBlueprintCreateRequest(request blueprint.CreateRequest) qovery.BlueprintCreateRequest {
	variables := make([]qovery.BlueprintVariableRequest, 0, len(request.Variables)+len(request.SecretVariables))
	for name, value := range request.Variables {
		variables = append(variables, qovery.BlueprintVariableRequest{Name: name, Value: value, IsSecret: qovery.PtrBool(false)})
	}
	for name, value := range request.SecretVariables {
		variables = append(variables, qovery.BlueprintVariableRequest{Name: name, Value: value, IsSecret: qovery.PtrBool(true)})
	}

	createRequest := qovery.BlueprintCreateRequest{
		Name:      request.Name,
		Tag:       request.Tag,
		Icon:      request.IconURI,
		Variables: variables,
	}
	if o := request.SpecOverrides; o != nil {
		createRequest.SpecOverrides = &qovery.BlueprintSpecOverrides{
			EngineVersion: o.EngineVersion,
			Credentials:   o.Credentials,
			Backend:       o.Backend,
			Timeout:       o.Timeout,
			Cpu:           o.CPU,
			Ram:           o.RAM,
			Storage:       o.Storage,
		}
	}
	return createRequest
}

// newQoveryBlueprintUpdateRequest builds an RFC 7396 merge-patch. The typed client fields
// cannot carry a null value, so the patch goes through AdditionalProperties, which the
// client serializes last and verbatim.
func newQoveryBlueprintUpdateRequest(request blueprint.UpdateRequest) qovery.BlueprintUpdateRequest {
	additionalProperties := map[string]interface{}{
		"variables": newBlueprintVariablesPatch(request),
	}
	if overrides := newBlueprintSpecOverridesPatch(request.SpecOverrides, request.PreviousSpecOverrides); len(overrides) > 0 {
		additionalProperties["spec_overrides"] = overrides
	}

	return qovery.BlueprintUpdateRequest{
		Name:                 request.Name,
		Tag:                  request.Tag,
		Icon:                 request.IconURI,
		AdditionalProperties: additionalProperties,
	}
}

func newBlueprintVariablesPatch(request blueprint.UpdateRequest) map[string]interface{} {
	patch := make(map[string]interface{}, len(request.Variables)+len(request.SecretVariables)+len(request.PreviousVariableNames))
	for _, name := range request.PreviousVariableNames {
		patch[name] = nil
	}
	for name, value := range request.Variables {
		patch[name] = qovery.BlueprintUpdateVariableValue{Value: value, IsSecret: qovery.PtrBool(false)}
	}
	for name, value := range request.SecretVariables {
		patch[name] = qovery.BlueprintUpdateVariableValue{Value: value, IsSecret: qovery.PtrBool(true)}
	}
	return patch
}

func newBlueprintSpecOverridesPatch(planned *blueprint.SpecOverrides, previous *blueprint.SpecOverrides) map[string]interface{} {
	if planned == nil {
		planned = &blueprint.SpecOverrides{}
	}
	if previous == nil {
		previous = &blueprint.SpecOverrides{}
	}

	patch := map[string]interface{}{}
	putStringField(patch, "engine_version", planned.EngineVersion, previous.EngineVersion)
	putStringField(patch, "credentials", planned.Credentials, previous.Credentials)
	putStringField(patch, "backend", planned.Backend, previous.Backend)
	putStringField(patch, "cpu", planned.CPU, previous.CPU)
	putStringField(patch, "ram", planned.RAM, previous.RAM)
	putStringField(patch, "storage", planned.Storage, previous.Storage)
	switch {
	case planned.Timeout != nil:
		patch["timeout"] = *planned.Timeout
	case previous.Timeout != nil:
		patch["timeout"] = nil
	}
	return patch
}

func putStringField(patch map[string]interface{}, key string, planned *string, previous *string) {
	switch {
	case planned != nil:
		patch[key] = *planned
	case previous != nil:
		patch[key] = nil
	}
}

func newDomainBlueprintFromQovery(details *qovery.BlueprintDetailsResponse, variables []qovery.BlueprintConfigurationVariable) (*blueprint.Blueprint, error) {
	if details == nil {
		return nil, errors.New("blueprint response is empty")
	}
	id, err := uuid.Parse(details.GetId())
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrInvalidBlueprintIDParam.Error())
	}
	environmentID, err := uuid.Parse(details.GetEnvironmentId())
	if err != nil {
		return nil, errors.Wrap(err, blueprint.ErrInvalidEnvironmentIDParam.Error())
	}

	domainVariables := make([]blueprint.Variable, 0, len(variables))
	for _, v := range variables {
		domainVariables = append(domainVariables, blueprint.Variable{
			Name:     v.GetName(),
			Value:    v.Value,
			IsSecret: v.GetIsSecret(),
		})
	}

	var latestDeployment *blueprint.Dispatch
	if d, ok := details.GetLatestDeploymentOk(); ok && d != nil {
		latestDeployment = &blueprint.Dispatch{
			ID:           d.GetId(),
			Status:       blueprint.DispatchStatus(d.GetStatus()),
			StartedAt:    d.GetStartedAt(),
			ErrorMessage: d.ErrorMessage.Get(),
		}
	}

	return &blueprint.Blueprint{
		ID:               id,
		EnvironmentID:    environmentID,
		Name:             details.GetName(),
		Tag:              details.GetTag(),
		CatalogURL:       details.GetCatalogUrl(),
		ServiceType:      blueprint.ServiceType(details.GetServiceType()),
		ServiceID:        details.ServiceId.Get(),
		LatestDeployment: latestDeployment,
		Variables:        domainVariables,
	}, nil
}

func newDomainCatalogEntriesFromQovery(catalog *qovery.BlueprintCatalogResponse) []blueprint.CatalogEntry {
	if catalog == nil {
		return nil
	}
	var entries []blueprint.CatalogEntry
	for _, item := range catalog.GetBlueprints() {
		for _, major := range item.GetMajorVersions() {
			entries = append(entries, blueprint.CatalogEntry{
				CatalogVersion: blueprint.CatalogVersion{
					Provider:       item.GetProvider(),
					ServiceFamily:  item.GetServiceFamily(),
					ServiceVersion: major.GetServiceVersion(),
				},
				LatestTag: major.GetLatestTag(),
			})
		}
	}
	return entries
}

// findServiceStatus returns nil when the environment does not report the service.
func findServiceStatus(statuses *qovery.EnvironmentStatuses, serviceType blueprint.ServiceType, serviceID string) (*blueprint.ServiceStatus, error) {
	var candidates func() []qovery.Status
	switch serviceType {
	case blueprint.ServiceTypeTerraform:
		candidates = statuses.GetTerraforms
	case blueprint.ServiceTypeHelm:
		candidates = statuses.GetHelms
	default:
		return nil, errors.Wrap(blueprint.ErrUnknownServiceType, string(serviceType))
	}
	if statuses == nil {
		return nil, nil
	}
	for _, s := range candidates() {
		if s.GetId() == serviceID {
			return &blueprint.ServiceStatus{
				State:              string(s.GetState()),
				LastDeploymentDate: s.LastDeploymentDate,
			}, nil
		}
	}
	return nil, nil
}
