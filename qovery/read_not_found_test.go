//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// newTestReadResponse builds a ReadResponse whose state holds a single non-null "id"
// attribute, so tests can assert whether handleReadNotFound removed it from state.
func newTestReadResponse() *resource.ReadResponse {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
		},
	}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}
	return &resource.ReadResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(objType, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "some-id")}),
			Schema: s,
		},
	}
}

func readErrorWithStatus(statusCode int) *apierrors.APIError {
	return apierrors.NewReadError(
		apierrors.APIResourceApplication,
		"some-id",
		&http.Response{StatusCode: statusCode},
		errors.New("boom"),
	)
}

func TestHandleReadNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		APIErr        *apierrors.APIError
		WantHandled   bool
		WantStateNull bool
		WantDiag      bool
	}{
		{
			TestName:    "nil error is not handled, Read continues",
			APIErr:      nil,
			WantHandled: false,
		},
		{
			TestName:      "404 removes the resource from state",
			APIErr:        readErrorWithStatus(http.StatusNotFound),
			WantHandled:   true,
			WantStateNull: true,
		},
		{
			TestName:      "403 is treated as not-found and removes from state",
			APIErr:        readErrorWithStatus(http.StatusForbidden),
			WantHandled:   true,
			WantStateNull: true,
		},
		{
			TestName:    "500 surfaces a diagnostic without removing from state",
			APIErr:      readErrorWithStatus(http.StatusInternalServerError),
			WantHandled: true,
			WantDiag:    true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			resp := newTestReadResponse()

			handled := handleReadNotFound(context.Background(), resp, tc.APIErr)

			assert.Equal(t, tc.WantHandled, handled)
			assert.Equal(t, tc.WantStateNull, resp.State.Raw.IsNull())
			assert.Equal(t, tc.WantDiag, resp.Diagnostics.HasError())
		})
	}
}
