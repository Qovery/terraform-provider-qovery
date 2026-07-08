//go:build unit && !integration

package qoveryapi

import (
	"testing"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDomainApiTokenFromCreateResponse(t *testing.T) {
	t.Parallel()

	organizationID := uuid.NewString()
	name := "my-token"
	description := "my description"
	token := "qov_secret_value"
	roleID := uuid.NewString()

	res := &qovery.OrganizationApiTokenCreate{
		Id:          uuid.NewString(),
		Name:        &name,
		Description: &description,
		Token:       &token,
		RoleId:      &roleID,
	}

	apiToken, err := newDomainApiTokenFromCreateResponse(organizationID, res)
	require.NoError(t, err)
	require.NotNil(t, apiToken)

	assert.Equal(t, res.Id, apiToken.ID.String())
	assert.Equal(t, organizationID, apiToken.OrganizationID.String())
	assert.Equal(t, name, apiToken.Name)
	assert.Equal(t, &description, apiToken.Description)
	assert.Equal(t, roleID, apiToken.RoleID)
	require.NotNil(t, apiToken.Token)
	assert.Equal(t, token, *apiToken.Token)
}

func TestNewDomainApiTokenFromListItem(t *testing.T) {
	t.Parallel()

	organizationID := uuid.NewString()
	name := "my-token"
	roleID := uuid.NewString()

	res := qovery.OrganizationApiToken{
		Id:     uuid.NewString(),
		Name:   &name,
		RoleId: &roleID,
	}

	apiToken, err := newDomainApiTokenFromListItem(organizationID, res)
	require.NoError(t, err)
	require.NotNil(t, apiToken)

	assert.Equal(t, res.Id, apiToken.ID.String())
	assert.Equal(t, organizationID, apiToken.OrganizationID.String())
	assert.Equal(t, name, apiToken.Name)
	assert.Nil(t, apiToken.Description)
	assert.Equal(t, roleID, apiToken.RoleID)
	assert.Nil(t, apiToken.Token)
}

func TestNewDomainApiTokenFromCreateResponseInvalidID(t *testing.T) {
	t.Parallel()

	res := &qovery.OrganizationApiTokenCreate{
		Id: "not-a-uuid",
	}

	apiToken, err := newDomainApiTokenFromCreateResponse(uuid.NewString(), res)
	assert.Error(t, err)
	assert.Nil(t, apiToken)
}
