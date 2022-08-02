package credentials

import (
	"github.com/go-playground/validator/v10"
)

// UpsertScalewayRequest represents the parameters needed to create & update Scaleway Credentials.
type UpsertScalewayRequest struct {
	Name              string `validate:"required"`
	ScalewayProjectID string `validate:"required"`
	ScalewayAccessKey string `validate:"required"`
	ScalewaySecretKey string `validate:"required"`
}

// Validate returns an error to tell whether the UpsertScalewayRequest is valid or not.
func (r UpsertScalewayRequest) Validate() error {
	return validator.New().Struct(r)
}

// IsValid returns a bool to tell whether the UpsertScalewayRequest is valid or not.
func (r UpsertScalewayRequest) IsValid() bool {
	return r.Validate() == nil
}
