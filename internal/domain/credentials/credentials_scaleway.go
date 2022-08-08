package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrInvalidUpsertScalewayRequest = errors.New("invalid credentials upsert scaleway request")
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
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertScalewayRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertScalewayRequest is valid or not.
func (r UpsertScalewayRequest) IsValid() bool {
	return r.Validate() == nil
}
