//go:build unit
// +build unit

package services_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func assertCreateCredentials(t *testing.T) *credentials.Credentials {
	creds, err := credentials.NewCredentials(credentials.NewCredentialsParams{
		CredentialsID:  gofakeit.UUID(),
		OrganizationID: gofakeit.UUID(),
		Name:           gofakeit.Name(),
	})
	require.NoError(t, err)
	require.NotNil(t, creds)
	require.NoError(t, creds.Validate())

	return creds
}

func assertEqualCredentials(t *testing.T, expected *credentials.Credentials, actual *credentials.Credentials) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.OrganizationID, actual.OrganizationID)
	assert.Equal(t, expected.Name, actual.Name)
}
