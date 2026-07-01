package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// handleReadNotFound centralizes not-found handling for client-layer resource Reads: on
// not-found (404/403, per apierrors.IsNotFound) it removes the resource from state so the
// next plan re-creates it; on any other error it adds a diagnostic. It reports whether Read
// should return early (true) or continue because apiErr is nil (false).
func handleReadNotFound(ctx context.Context, resp *resource.ReadResponse, apiErr *apierrors.APIError) bool {
	if apiErr == nil {
		return false
	}
	if apierrors.IsNotFound(apiErr) {
		resp.State.RemoveResource(ctx)
		return true
	}
	resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
	return true
}
