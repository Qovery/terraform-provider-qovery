package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// ErrInvalidUpsertGcpRequest is returned if a GCP Credentials upsert request is invalid.
var ErrInvalidUpsertGcpRequest = errors.New("invalid credentials upsert gcp request")

// GcpServiceAccountKeyCredentials holds the static service-account JSON key auth mode.
type GcpServiceAccountKeyCredentials struct {
	GcpCredentials string `validate:"required"`
}

// GcpWorkloadIdentityCredentials holds the Workload Identity Federation auth mode.
type GcpWorkloadIdentityCredentials struct {
	ServiceAccountEmail              string `validate:"required"`
	WorkloadIdentityProviderResource string `validate:"required"`
}

// UpsertGcpRequest represents the parameters needed to create & update GCP Credentials.
// Exactly one of ServiceAccountKey or WorkloadIdentity must be set.
type UpsertGcpRequest struct {
	Name              string `validate:"required"`
	ServiceAccountKey *GcpServiceAccountKeyCredentials
	WorkloadIdentity  *GcpWorkloadIdentityCredentials
}

// Validate returns an error to tell whether the UpsertGcpRequest is valid or not.
func (r UpsertGcpRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertGcpRequest.Error())
	}

	if r.ServiceAccountKey == nil && r.WorkloadIdentity == nil {
		return errors.Wrap(errors.New("either ServiceAccountKey or WorkloadIdentity must be provided"), ErrInvalidUpsertGcpRequest.Error())
	}
	if r.ServiceAccountKey != nil && r.WorkloadIdentity != nil {
		return errors.Wrap(errors.New("only one of ServiceAccountKey or WorkloadIdentity must be provided"), ErrInvalidUpsertGcpRequest.Error())
	}

	if r.ServiceAccountKey != nil {
		if err := validator.New().Struct(r.ServiceAccountKey); err != nil {
			return errors.Wrap(err, ErrInvalidUpsertGcpRequest.Error())
		}
	}
	if r.WorkloadIdentity != nil {
		if err := validator.New().Struct(r.WorkloadIdentity); err != nil {
			return errors.Wrap(err, ErrInvalidUpsertGcpRequest.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertGcpRequest is valid or not.
func (r UpsertGcpRequest) IsValid() bool {
	return r.Validate() == nil
}
