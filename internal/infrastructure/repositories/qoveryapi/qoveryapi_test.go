package qoveryapi_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/qoveryapi"
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

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		HTTPClient    *http.Client
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_http_client",
			ExpectedError: qoveryapi.ErrInvalidHTTPClient,
		},
		{
			TestName:   "success_with_http_client",
			HTTPClient: &http.Client{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			qoveryAPI, err := qoveryapi.New(qoveryapi.WithHTTPClient(tc.HTTPClient))
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, qoveryAPI)
				return
			}

			assert.NoError(t, err)
			assert.Same(t, tc.HTTPClient, qoveryAPI.Client.GetConfig().HTTPClient)
		})
	}
}
