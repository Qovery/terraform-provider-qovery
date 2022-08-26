package services_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		TestName      string
		Configs       []services.Configuration
		ExpectedError error
	}{
		{
			TestName:      "fail_without_configuration",
			ExpectedError: services.ErrMissingConfiguration,
		},
		{
			TestName: "success_with_configuration",
			Configs: []services.Configuration{
				services.WithQoveryRepository(gofakeit.Name(), gofakeit.Name()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			svc, err := services.New(tc.Configs...)
			if tc.ExpectedError != nil {
				assert.Nil(t, svc)
				assert.ErrorContains(t, err, tc.ExpectedError.Error())

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, svc)
			assert.NotNil(t, svc.Organization)
			assert.NotNil(t, svc.CredentialsAws)
			assert.NotNil(t, svc.CredentialsScaleway)
			assert.NotNil(t, svc.Project)
		})
	}
}
