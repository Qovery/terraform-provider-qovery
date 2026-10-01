//go:build unit && !integration
// +build unit,!integration

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDatabaseAPI serves the database endpoints and records every edit request body.
type fakeDatabaseAPI struct {
	mu          sync.Mutex
	description string
	readStatus  int
	edits       []map[string]any
}

func (f *fakeDatabaseAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.URL.Path != "/database/db-1" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if f.readStatus != 0 {
			w.WriteHeader(f.readStatus)
			return
		}
	case http.MethodPut:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.edits = append(f.edits, body)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           "db-1",
		"created_at":   "2026-09-28T00:00:00Z",
		"name":         "db",
		"description":  f.description,
		"type":         "REDIS",
		"version":      "6.2",
		"mode":         "CONTAINER",
		"icon_uri":     "app://qovery-console/database",
		"environment":  map[string]any{"id": "env-1"},
		"service_type": "DATABASE",
	})
}

func TestClient_editDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName          string
		RemoteDescription string
		ReadStatus        int
		ExpectError       bool
	}{
		{
			TestName:          "description_set_in_the_console_is_resent",
			RemoteDescription: "set in the Console",
		},
		{
			TestName:    "read_failure_skips_the_edit",
			ReadStatus:  http.StatusInternalServerError,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &fakeDatabaseAPI{description: tc.RemoteDescription, readStatus: tc.ReadStatus}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			c := New("token", "test", server.URL)
			request := qovery.DatabaseEditRequest{Name: qovery.PtrString("planned-name")}
			database, apiErr := c.editDatabase(context.Background(), "db-1", request)

			if tc.ExpectError {
				require.NotNil(t, apiErr)
				assert.Empty(t, api.edits)
				return
			}
			require.Nil(t, apiErr)
			require.NotNil(t, database)
			require.Len(t, api.edits, 1)
			assert.Equal(t, tc.RemoteDescription, api.edits[0]["description"])
			assert.Equal(t, "planned-name", api.edits[0]["name"])
		})
	}
}
