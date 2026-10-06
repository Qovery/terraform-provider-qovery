//go:build unit && !integration

package qoveryapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

type organizationAPICall struct {
	method string
	path   string
	body   string
}

type organizationAPIAnswer struct {
	status int
	body   string
}

// newOrganizationQoveryAPIServing answers each HTTP method with its entry in answers and records every call.
func newOrganizationQoveryAPIServing(t *testing.T, answers map[string]organizationAPIAnswer) (organization.Repository, func() []organizationAPICall) {
	t.Helper()
	var (
		mu    sync.Mutex
		calls []organizationAPICall
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		mu.Lock()
		calls = append(calls, organizationAPICall{method: r.Method, path: r.URL.Path, body: string(body)})
		mu.Unlock()
		answer, ok := answers[r.Method]
		if !ok {
			answer = organizationAPIAnswer{status: http.StatusMethodNotAllowed, body: `{"message":"unexpected method"}`}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(answer.status)
		_, _ = w.Write([]byte(answer.body))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	repo, err := newOrganizationQoveryAPI(qovery.NewAPIClient(cfg))
	require.NoError(t, err)
	return repo, func() []organizationAPICall {
		mu.Lock()
		defer mu.Unlock()
		return append([]organizationAPICall(nil), calls...)
	}
}

func TestOrganizationQoveryAPIUpdate(t *testing.T) {
	t.Parallel()
	organizationID := uuid.NewString()
	organizationPath := "/organization/" + organizationID
	editedOrganization := fmt.Sprintf(`{"id":%q,"created_at":"2026-01-01T00:00:00Z","name":"new name","description":"new description","plan":"ENTERPRISE"}`, organizationID)
	request := organization.UpdateRequest{
		Name:        "new name",
		Description: new("new description"),
	}

	testCases := []struct {
		name         string
		unmanaged    string
		wantEditBody string
	}{
		{
			name:      "resends the unmanaged fields of the current organization",
			unmanaged: `"website_url":"https://www.example.com","logo_url":"https://example.com/logo.png","icon_url":"https://example.com/icon.png","admin_emails":["admin@example.com"]`,
			wantEditBody: `{"name":"new name","description":"new description",` +
				`"website_url":"https://www.example.com","logo_url":"https://example.com/logo.png","icon_url":"https://example.com/icon.png","admin_emails":["admin@example.com"]}`,
		},
		{
			name:         "resends null unmanaged fields as null",
			unmanaged:    `"website_url":null,"logo_url":null,"icon_url":null,"admin_emails":[]`,
			wantEditBody: `{"name":"new name","description":"new description","website_url":null,"logo_url":null,"icon_url":null,"admin_emails":[]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			currentOrganization := fmt.Sprintf(`{"id":%q,"created_at":"2026-01-01T00:00:00Z","name":"current name","description":"current description","plan":"ENTERPRISE",%s}`, organizationID, tc.unmanaged)
			repo, calls := newOrganizationQoveryAPIServing(t, map[string]organizationAPIAnswer{
				http.MethodGet: {status: http.StatusOK, body: currentOrganization},
				http.MethodPut: {status: http.StatusOK, body: editedOrganization},
			})

			orga, err := repo.Update(context.Background(), organizationID, request)

			require.NoError(t, err)
			assert.Equal(t, "new name", orga.Name)
			recorded := calls()
			require.Len(t, recorded, 2)
			assert.Equal(t, organizationAPICall{method: http.MethodGet, path: organizationPath}, recorded[0])
			assert.Equal(t, http.MethodPut, recorded[1].method)
			assert.Equal(t, organizationPath, recorded[1].path)
			assert.JSONEq(t, tc.wantEditBody, recorded[1].body)
		})
	}

	t.Run("read failure sends no edit", func(t *testing.T) {
		t.Parallel()
		repo, calls := newOrganizationQoveryAPIServing(t, map[string]organizationAPIAnswer{
			http.MethodGet: {status: http.StatusInternalServerError, body: `{"message":"stub error"}`},
		})

		orga, err := repo.Update(context.Background(), organizationID, request)

		var apiErr *apierrors.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Nil(t, orga)
		assert.Equal(t, []organizationAPICall{{method: http.MethodGet, path: organizationPath}}, calls())
	})
}
