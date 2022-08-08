package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestNewCredentials(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        credentials.NewCredentialsParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_credentials_id",
			Params: credentials.NewCredentialsParams{
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
			},
			ExpectedError: credentials.ErrInvalidCredentialsID,
		},
		{
			TestName: "fail_with_invalid_organization_id",
			Params: credentials.NewCredentialsParams{
				CredentialsID: gofakeit.UUID(),
				Name:          gofakeit.Name(),
			},
			ExpectedError: credentials.ErrInvalidCredentialsOrganizationID,
		},
		{
			TestName: "fail_with_invalid_name",
			Params: credentials.NewCredentialsParams{
				CredentialsID:  gofakeit.UUID(),
				OrganizationID: gofakeit.UUID(),
			},
			ExpectedError: credentials.ErrInvalidCredentialsName,
		},
		{
			TestName: "success",
			Params: credentials.NewCredentialsParams{
				CredentialsID:  gofakeit.UUID(),
				OrganizationID: gofakeit.UUID(),
				Name:           gofakeit.Name(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := credentials.NewCredentials(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.Equal(t, tc.Params.CredentialsID, creds.ID.String())
			assert.Equal(t, tc.Params.OrganizationID, creds.OrganizationID.String())
			assert.Equal(t, tc.Params.Name, creds.Name)
		})
	}
}
