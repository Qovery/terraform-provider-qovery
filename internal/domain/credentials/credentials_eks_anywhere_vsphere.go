package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var ErrInvalidUpsertEksAnywhereVsphereRequest = errors.New("invalid credentials upsert eks anywhere vsphere request")

type VsphereStaticCredentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

type VsphereRoleCredentials struct {
	RoleArn string `validate:"required"`
}

// UpsertEksAnywhereVsphereRequest represents the parameters needed to create & update EKS Anywhere vSphere Credentials.
type UpsertEksAnywhereVsphereRequest struct {
	Name              string `validate:"required"`
	VsphereUser       string `validate:"required"`
	VspherePassword   string `validate:"required"`
	StaticCredentials *VsphereStaticCredentials
	RoleCredentials   *VsphereRoleCredentials
}

// Validate returns an error to tell whether the UpsertEksAnywhereVsphereRequest is valid or not.
func (r UpsertEksAnywhereVsphereRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertEksAnywhereVsphereRequest.Error())
	}

	if r.StaticCredentials == nil && r.RoleCredentials == nil {
		return errors.Wrap(errors.New("either StaticCredentials or RoleCredentials must be provided"), ErrInvalidUpsertEksAnywhereVsphereRequest.Error())
	}

	if r.StaticCredentials != nil && r.RoleCredentials != nil {
		return errors.Wrap(errors.New("StaticCredentials and RoleCredentials are mutually exclusive"), ErrInvalidUpsertEksAnywhereVsphereRequest.Error())
	}

	if r.StaticCredentials != nil {
		if err := validator.New().Struct(r.StaticCredentials); err != nil {
			return errors.Wrap(err, ErrInvalidUpsertEksAnywhereVsphereRequest.Error())
		}
	}

	if r.RoleCredentials != nil {
		if err := validator.New().Struct(r.RoleCredentials); err != nil {
			return errors.Wrap(err, ErrInvalidUpsertEksAnywhereVsphereRequest.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertEksAnywhereVsphereRequest is valid or not.
func (r UpsertEksAnywhereVsphereRequest) IsValid() bool {
	return r.Validate() == nil
}
