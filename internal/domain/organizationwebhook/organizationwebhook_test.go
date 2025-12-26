package organizationwebhook_test

import (
	"testing"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organizationwebhook"
	"github.com/stretchr/testify/assert"
)

func TestKind_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Params      organizationwebhook.OrganizationWebhook
		ExpectedErr error
	}{
		{
			TestName:    "fail_with_invalid_kind",
			Params:      organizationwebhook.OrganizationWebhook{Kind: organizationwebhook.Kind("INVALID_KIND")},
			ExpectedErr: organizationwebhook.ErrInvalidKindOrganizationWebhook,
		},
		{
			TestName:    "success_with_standard_kind",
			Params:      organizationwebhook.OrganizationWebhook{Kind: organizationwebhook.KindStandard},
			ExpectedErr: nil,
		},
		{
			TestName:    "success_with_slack_kind",
			Params:      organizationwebhook.OrganizationWebhook{Kind: organizationwebhook.KindSlack},
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			err := tc.Params.Kind.Validate()
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Params      organizationwebhook.OrganizationWebhook
		ExpectedErr error
	}{
		{
			TestName:    "fail_with_invalid_event",
			Params:      organizationwebhook.OrganizationWebhook{Events: []organizationwebhook.Event{"INVALID_EVENT"}},
			ExpectedErr: organizationwebhook.ErrInvalidEventOrganizationWebhook,
		},
		{
			TestName:    "success_with_valid_events",
			Params:      organizationwebhook.OrganizationWebhook{Events: []organizationwebhook.Event{organizationwebhook.EventDeploymentStarted, organizationwebhook.EventDeploymentSuccessful}},
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			var err error
			for _, event := range tc.Params.Events {
				err = event.Validate()
				if err != nil {
					break
				}
			}
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURL_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Params      organizationwebhook.OrganizationWebhook
		ExpectedErr error
	}{
		{
			TestName:    "fail_with_empty_url",
			Params:      organizationwebhook.OrganizationWebhook{URL: "", Kind: organizationwebhook.KindStandard, Events: []organizationwebhook.Event{organizationwebhook.EventDeploymentStarted}},
			ExpectedErr: organizationwebhook.ErrInvalidOrganizationWebhookURLParam,
		},
		{
			TestName:    "fail_with_invalid_url",
			Params:      organizationwebhook.OrganizationWebhook{URL: "invalid-url", Kind: organizationwebhook.KindStandard, Events: []organizationwebhook.Event{organizationwebhook.EventDeploymentStarted}},
			ExpectedErr: organizationwebhook.ErrInvalidOrganizationWebhookURLParam,
		},
		{
			TestName:    "success_with_valid_url",
			Params:      organizationwebhook.OrganizationWebhook{URL: "https://example.com/webhook", Kind: organizationwebhook.KindStandard, Events: []organizationwebhook.Event{organizationwebhook.EventDeploymentStarted}},
			ExpectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			err := tc.Params.Validate()
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
