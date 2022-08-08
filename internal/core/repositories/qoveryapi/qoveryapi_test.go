package qoveryapi_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/qoveryapi"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Configs       []qoveryapi.Configuration
		ExpectedError error
	}{
		{
			TestName: "success_without_configuration",
			Configs:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			qoveryAPI, err := qoveryapi.New(tc.Configs...)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, qoveryAPI)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, qoveryAPI)
			assert.NotNil(t, qoveryAPI.CredentialsAws)
			assert.NotNil(t, qoveryAPI.CredentialsScaleway)
		})
	}
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		UserAgent     string
		ExpectedError error
	}{
		{
			TestName:      "fail_with_invalid_user_agent",
			ExpectedError: qoveryapi.ErrInvalidUserAgent,
		},
		{
			TestName:  "success_with_valid_user_agent",
			UserAgent: gofakeit.UserAgent(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			qoveryAPI, err := qoveryapi.New(qoveryapi.WithUserAgent(tc.UserAgent))
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, qoveryAPI)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, qoveryAPI)
		})
	}
}

func TestWithQoveryAPIToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		APIToken      string
		ExpectedError error
	}{
		{
			TestName:      "fail_with_invalid_api_token",
			ExpectedError: qoveryapi.ErrInvalidQoveryAPIToken,
		},
		{
			TestName: "success_with_valid_api_token",
			APIToken: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			qoveryAPI, err := qoveryapi.New(qoveryapi.WithQoveryAPIToken(tc.APIToken))
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, qoveryAPI)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, qoveryAPI)
		})
	}
}
