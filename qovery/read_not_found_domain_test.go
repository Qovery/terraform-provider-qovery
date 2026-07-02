//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"net/http"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	domainapierrors "github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func domainReadErrorWithStatus(statusCode int) error {
	return domainapierrors.NewReadAPIError(
		domainapierrors.APIResourceOrganization,
		"some-id",
		&http.Response{StatusCode: statusCode},
		pkgerrors.New("boom"),
	)
}

func TestHandleDomainReadNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Err           error
		WantStateNull bool
		WantDiag      bool
	}{
		{
			TestName: "nil error is not handled, Read continues",
			Err:      nil,
		},
		{
			TestName:      "raw domain 404 removes the resource from state",
			Err:           domainReadErrorWithStatus(http.StatusNotFound),
			WantStateNull: true,
		},
		{
			TestName:      "domain 404 wrapped by the service layer still removes from state",
			Err:           pkgerrors.Wrap(domainReadErrorWithStatus(http.StatusNotFound), "failed to get organization"),
			WantStateNull: true,
		},
		{
			TestName:      "403 is treated as not-found and removes from state",
			Err:           pkgerrors.Wrap(domainReadErrorWithStatus(http.StatusForbidden), "failed to get organization"),
			WantStateNull: true,
		},
		{
			TestName: "plain error surfaces a diagnostic",
			Err:      pkgerrors.New("some other failure"),
			WantDiag: true,
		},
		{
			TestName: "500 surfaces a diagnostic without removing from state",
			Err:      pkgerrors.Wrap(domainReadErrorWithStatus(http.StatusInternalServerError), "failed to get organization"),
			WantDiag: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			resp := newTestReadResponse()

			handled := handleDomainReadNotFound(context.Background(), resp, tc.Err, "Error on test read")

			assert.Equal(t, tc.Err != nil, handled)
			assert.Equal(t, tc.WantStateNull, resp.State.Raw.IsNull())
			assert.Equal(t, tc.WantDiag, resp.Diagnostics.HasError())
		})
	}
}
