package qovery

import (
	"context"
	"errors"

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
// Reads whose service returns a plain error. Not-found is detected either by a typed
// sentinel (errors.Is against the given sentinels) or by an internal/domain/apierrors
// *APIError anywhere in the wrap chain reporting IsNotFound (404, 403, and 400+"exist",
// per that package). Services wrap repository errors with pkg/errors, so detection uses
// errors.As rather than the package's cast-only IsErrNotFound. On not-found the resource
// is removed from state so the next plan re-creates it; any other error is surfaced as a
// diagnostic under the given summary. It reports whether Read should return early (true)
// or continue because err is nil (false).
func handleDomainReadNotFound(ctx context.Context, resp *resource.ReadResponse, err error, summary string, sentinels ...error) bool {
	if err == nil {
		return false
	}
	for _, sentinel := range sentinels {
		if errors.Is(err, sentinel) {
			resp.State.RemoveResource(ctx)
			return true
		}
	}
	var apiErr *domainapierrors.APIError
	if errors.As(err, &apiErr) && apiErr.IsNotFound() {
		resp.State.RemoveResource(ctx)
		return true
	}
	resp.Diagnostics.AddError(summary, err.Error())
	return true
}
