package secret_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        secret.NewSecretParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_secret_id",
			Params: secret.NewSecretParams{
				Scope: variable.ScopeApplication.String(),
				Key:   gofakeit.Name(),
				Type:  "VALUE",
			},
			ExpectedError: secret.ErrInvalidSecretIDParam,
		},
		{
			TestName: "fail_with_invalid_key",
			Params: secret.NewSecretParams{
				SecretID: gofakeit.UUID(),
				Scope:    variable.ScopeApplication.String(),
				Type:     "VALUE",
			},
			ExpectedError: secret.ErrInvalidKeyParam,
		},
		{
			TestName: "fail_with_invalid_scope",
			Params: secret.NewSecretParams{
				SecretID: gofakeit.UUID(),
				Key:      gofakeit.Name(),
				Type:     "VALUE",
			},
			ExpectedError: secret.ErrInvalidScopeParam,
		},
		{
			TestName: "success",
			Params: secret.NewSecretParams{
				SecretID: gofakeit.UUID(),
				Scope:    variable.ScopeApplication.String(),
				Key:      gofakeit.Name(),
				Type:     "VALUE",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			v, err := secret.NewSecret(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, v)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, v)
			assert.True(t, v.IsValid())
			assert.Equal(t, tc.Params.SecretID, v.ID.String())
			assert.Equal(t, tc.Params.Key, v.Key)
			assert.Equal(t, tc.Params.Type, v.Type)
		})
	}
}
