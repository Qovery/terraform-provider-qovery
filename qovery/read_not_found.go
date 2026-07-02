package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	domainapierrors "github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
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

// handleDomainReadNotFound is the domain-service twin of handleReadNotFound, for resource
// Reads whose service returns a plain error. Not-found is detected by apierrors.IsErrNotFound
// (an internal/domain/apierrors *APIError anywhere in the wrap chain reporting 404, 403, or
// 400+"exist", per that package). On not-found the resource is removed from state so the next
// plan re-creates it; any other error is surfaced as a diagnostic under the given summary.
// It reports whether Read should return early (true) or continue because err is nil (false).
func handleDomainReadNotFound(ctx context.Context, resp *resource.ReadResponse, err error, summary string) bool {
	if err == nil {
		return false
	}
	if domainapierrors.IsErrNotFound(err) {
		resp.State.RemoveResource(ctx)
		return true
	}
	resp.Diagnostics.AddError(summary, err.Error())
	return true
}
