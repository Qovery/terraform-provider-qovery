//go:build unit && !integration
// +build unit,!integration

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestApplicationExternalSecrets_JSONTargetKey(t *testing.T) {
	t.Parallel()
	for _, selector := range []*string{nil, new(`accounts.0.database\.password`)} {
		for _, operation := range []string{"create", "update", "list"} {
			name := operation
			if selector != nil {
				name += "_json"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				selectorJSON, err := json.Marshal(selector)
				require.NoError(t, err)
				response := fmt.Sprintf(`{"id":"00000000-0000-0000-0000-000000000001","key":"PASSWORD","value":"/prod/db","scope":"APPLICATION","variable_type":"EXTERNAL_SECRET","created_at":"2026-11-03T00:00:00Z","is_secret":false,"json_target_key":%s}`, selectorJSON)
				calls := make(chan map[string]json.RawMessage, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.Method == http.MethodGet {
						_, _ = fmt.Fprintf(w, `{"results":[%s]}`, response)
						return
					}
					var payload map[string]json.RawMessage
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
					calls <- payload
					_, _ = w.Write([]byte(response))
				}))
				defer server.Close()
				client := New("token", "test", server.URL, newTestHTTPClient())
				request := variable.ExternalSecretUpsertRequest{Key: "PASSWORD", Reference: "/prod/db", SecretManagerAccessId: "manager", JSONTargetKey: selector}
				if operation == "list" {
					secrets, files, apiErr := client.getApplicationExternalSecretsAndFiles(context.Background(), "application-id")
					require.Nil(t, apiErr)
					require.Len(t, secrets, 1)
					assert.Empty(t, files)
					assert.Equal(t, selector, secrets[0].JSONTargetKey)
					return
				}
				diff := variable.ExternalSecretDiffRequest{}
				if operation == "create" {
					diff.Create = []variable.ExternalSecretDiffCreateRequest{{ExternalSecretUpsertRequest: request}}
				} else {
					diff.Update = []variable.ExternalSecretDiffUpdateRequest{{ExternalSecretUpsertRequest: request, VariableID: "variable-id"}}
				}
				require.Nil(t, client.updateApplicationExternalSecrets(context.Background(), "application-id", diff))
				select {
				case payload := <-calls:
					require.Contains(t, payload, "json_target_key")
					assert.JSONEq(t, string(selectorJSON), string(payload["json_target_key"]))
				default:
					t.Fatal("expected a variable write")
				}
			})
		}
	}
}
