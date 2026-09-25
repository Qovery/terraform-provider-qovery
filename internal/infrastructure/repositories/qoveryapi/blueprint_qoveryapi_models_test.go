//go:build unit && !integration

package qoveryapi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

func blueprintPtr[T any](v T) *T { return &v }

func TestNewQoveryBlueprintCreateRequest(t *testing.T) {
	t.Parallel()
	request := blueprint.CreateRequest{UpsertRequest: blueprint.UpsertRequest{
		Name:            "my-db",
		Tag:             "aws/postgres/17/1.0.0",
		IconURI:         "app://qovery-console/terraform",
		Variables:       map[string]string{"instance_type": "db.t3.micro", "allocated_storage": "20"},
		SecretVariables: map[string]string{"api_key": "test-value-1"},
		SpecOverrides:   &blueprint.SpecOverrides{CPU: blueprintPtr("500m"), Timeout: blueprintPtr(int32(600))},
	}}

	body, err := json.Marshal(newQoveryBlueprintCreateRequest(request))
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"name": "my-db",
		"tag": "aws/postgres/17/1.0.0",
		"icon": "app://qovery-console/terraform",
		"variables": [
			{"name": "allocated_storage", "value": "20", "is_secret": false},
			{"name": "instance_type", "value": "db.t3.micro", "is_secret": false},
			{"name": "api_key", "value": "test-value-1", "is_secret": true}
		],
		"spec_overrides": {"cpu": "500m", "timeout": 600}
	}`, string(body))
}

func TestNewQoveryBlueprintUpdateRequest(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		request  blueprint.UpdateRequest
		expected string
	}{
		{
			name: "removed variables and overrides are sent as null",
			request: blueprint.UpdateRequest{
				UpsertRequest: blueprint.UpsertRequest{
					Name:            "my-db",
					Tag:             "aws/postgres/17/1.0.1",
					IconURI:         "app://qovery-console/terraform",
					Variables:       map[string]string{"instance_type": "db.t3.small"},
					SecretVariables: map[string]string{"api_key": "test-value-2"},
					SpecOverrides:   &blueprint.SpecOverrides{RAM: blueprintPtr("1Gi")},
				},
				PreviousVariableNames: []string{"instance_type", "storage_gb", "api_key"},
				PreviousSpecOverrides: &blueprint.SpecOverrides{CPU: blueprintPtr("500m"), Timeout: blueprintPtr(int32(600))},
			},
			expected: `{
				"name": "my-db",
				"tag": "aws/postgres/17/1.0.1",
				"icon": "app://qovery-console/terraform",
				"variables": {
					"instance_type": {"value": "db.t3.small", "is_secret": false},
					"storage_gb": null,
					"api_key": {"value": "test-value-2", "is_secret": true}
				},
				"spec_overrides": {"ram": "1Gi", "cpu": null, "timeout": null}
			}`,
		},
		{
			name: "no overrides before nor after omits spec_overrides",
			request: blueprint.UpdateRequest{
				UpsertRequest: blueprint.UpsertRequest{Name: "my-db", Tag: "aws/postgres/17/1.0.1", IconURI: "app://qovery-console/terraform"},
			},
			expected: `{
				"name": "my-db",
				"tag": "aws/postgres/17/1.0.1",
				"icon": "app://qovery-console/terraform",
				"variables": {}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body, err := json.Marshal(newQoveryBlueprintUpdateRequest(tc.request))
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(body))
		})
	}
}

func TestNewDomainBlueprintFromQovery(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	environmentID := uuid.New()
	serviceID := uuid.NewString()
	details := qovery.BlueprintDetailsResponse{
		Id:            id.String(),
		Name:          "my-db",
		CatalogUrl:    "https://catalog/aws/postgres",
		Tag:           "aws/postgres/17/1.0.0",
		EnvironmentId: environmentID.String(),
		ServiceType:   "TERRAFORM",
		ServiceId:     *qovery.NewNullableString(&serviceID),
		LatestDeployment: *qovery.NewNullableBlueprintDetailsResponseLatestDeployment(&qovery.BlueprintDetailsResponseLatestDeployment{
			Id:           "dispatch-1",
			Status:       "FAILED",
			ErrorMessage: *qovery.NewNullableString(blueprintPtr("boom")),
		}),
	}
	variables := []qovery.BlueprintConfigurationVariable{
		{Name: "instance_type", Value: blueprintPtr("db.t3.micro"), IsSecret: false},
		{Name: "api_key", IsSecret: true},
	}

	bp, err := newDomainBlueprintFromQovery(&details, variables)
	require.NoError(t, err)
	assert.Equal(t, &blueprint.Blueprint{
		ID:               id,
		EnvironmentID:    environmentID,
		Name:             "my-db",
		Tag:              "aws/postgres/17/1.0.0",
		CatalogURL:       "https://catalog/aws/postgres",
		ServiceType:      blueprint.ServiceTypeTerraform,
		ServiceID:        &serviceID,
		LatestDeployment: &blueprint.Dispatch{ID: "dispatch-1", Status: blueprint.DispatchStatusFailed, ErrorMessage: blueprintPtr("boom")},
		Variables: []blueprint.Variable{
			{Name: "instance_type", Value: blueprintPtr("db.t3.micro")},
			{Name: "api_key", IsSecret: true},
		},
	}, bp)

	details.LatestDeployment = *qovery.NewNullableBlueprintDetailsResponseLatestDeployment(nil)
	details.ServiceId = *qovery.NewNullableString(nil)
	bp, err = newDomainBlueprintFromQovery(&details, nil)
	require.NoError(t, err)
	assert.Nil(t, bp.LatestDeployment)
	assert.Nil(t, bp.ServiceID)

	_, err = newDomainBlueprintFromQovery(nil, nil)
	assert.Error(t, err)
}

func TestFindServiceStatus(t *testing.T) {
	t.Parallel()
	serviceID := uuid.NewString()
	deployedAt := time.Date(2026, 9, 24, 14, 28, 36, 0, time.UTC)
	statuses := &qovery.EnvironmentStatuses{
		Terraforms: []qovery.Status{{Id: serviceID, State: "DEPLOYED", LastDeploymentDate: &deployedAt}},
		Helms:      []qovery.Status{{Id: uuid.NewString(), State: "DEPLOYING"}},
	}

	found, err := findServiceStatus(statuses, blueprint.ServiceTypeTerraform, serviceID)
	require.NoError(t, err)
	assert.Equal(t, &blueprint.ServiceStatus{State: "DEPLOYED", LastDeploymentDate: &deployedAt}, found)

	found, err = findServiceStatus(statuses, blueprint.ServiceTypeHelm, serviceID)
	require.NoError(t, err)
	assert.Nil(t, found)

	_, err = findServiceStatus(statuses, blueprint.ServiceType("JOB"), serviceID)
	assert.ErrorIs(t, err, blueprint.ErrUnknownServiceType)

	_, err = findServiceStatus(nil, blueprint.ServiceType("JOB"), serviceID)
	assert.ErrorIs(t, err, blueprint.ErrUnknownServiceType)

	found, err = findServiceStatus(nil, blueprint.ServiceTypeHelm, serviceID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestNewDomainCatalogEntriesFromQovery(t *testing.T) {
	t.Parallel()
	catalog := &qovery.BlueprintCatalogResponse{Blueprints: []qovery.BlueprintItem{{
		Provider:      "AWS",
		ServiceFamily: "postgres",
		MajorVersions: []qovery.BlueprintMajorVersion{
			{ServiceVersion: "16", LatestTag: "AWS/postgres/16/4.1.0"},
			{ServiceVersion: "17", LatestTag: "AWS/postgres/17/4.1.0"},
		},
	}}}

	assert.Equal(t, []blueprint.CatalogEntry{
		{CatalogVersion: blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "16"}, LatestTag: "AWS/postgres/16/4.1.0"},
		{CatalogVersion: blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "17"}, LatestTag: "AWS/postgres/17/4.1.0"},
	}, newDomainCatalogEntriesFromQovery(catalog))
	assert.Nil(t, newDomainCatalogEntriesFromQovery(nil))
}
