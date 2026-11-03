//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestExternalSecretsQoveryAPI_JSONTargetKey(t *testing.T) {
	t.Parallel()
	for _, scope := range []qovery.APIVariableScopeEnum{
		qovery.APIVARIABLESCOPEENUM_APPLICATION, qovery.APIVARIABLESCOPEENUM_CONTAINER,
		qovery.APIVARIABLESCOPEENUM_JOB, qovery.APIVARIABLESCOPEENUM_HELM,
		qovery.APIVARIABLESCOPEENUM_TERRAFORM, qovery.APIVARIABLESCOPEENUM_ENVIRONMENT,
	} {
		for _, selector := range []*string{nil, new(`accounts.0.database\.password`)} {
			for _, operation := range []string{"create", "update", "list"} {
				name := string(scope) + "_" + operation
				if selector != nil {
					name += "_json"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					selectorJSON, err := json.Marshal(selector)
					require.NoError(t, err)
					body := fmt.Sprintf(`{"id":"00000000-0000-0000-0000-000000000001","key":"PASSWORD","value":"/prod/db","scope":%q,"variable_type":"EXTERNAL_SECRET","created_at":"2026-11-03T00:00:00Z","is_secret":false,"secret_manager_access_id":"00000000-0000-0000-0000-000000000002","json_target_key":%s}`, scope, selectorJSON)
					if operation == "list" {
						body = `{"results":[` + body + `]}`
					}
					api, calls := newDeploymentActionAPIClient(t, http.StatusOK, body)
					repo, err := newExternalSecretsQoveryAPI(api, scope, apierrors.APIResourceEnvironmentExternalSecret)
					require.NoError(t, err)
					request := variable.ExternalSecretUpsertRequest{Key: "PASSWORD", Reference: "/prod/db", SecretManagerAccessId: "00000000-0000-0000-0000-000000000002", JSONTargetKey: selector}
					var result *variable.ExternalSecret
					switch operation {
					case "create":
						result, err = repo.Create(context.Background(), "parent-id", request)
					case "update":
						result, err = repo.Update(context.Background(), "variable-id", request)
					case "list":
						var values variable.ExternalSecrets
						values, err = repo.List(context.Background(), "parent-id")
						require.Len(t, values, 1)
						result = &values[0]
					}
					require.NoError(t, err)
					assert.Equal(t, selector, result.JSONTargetKey)
					require.Len(t, calls(), 1)
					if operation != "list" {
						var payload map[string]json.RawMessage
						require.NoError(t, json.Unmarshal([]byte(calls()[0].body), &payload))
						require.Contains(t, payload, "json_target_key")
						assert.JSONEq(t, string(selectorJSON), string(payload["json_target_key"]))
					}
				})
			}
		}
	}
}
