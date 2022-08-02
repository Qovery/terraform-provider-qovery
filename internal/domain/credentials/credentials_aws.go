package credentials

import (
	"github.com/go-playground/validator/v10"
)

// UpsertAwsRequest represents the parameters needed to create & update AWS Credentials.
type UpsertAwsRequest struct {
	Name            string `validate:"required"`
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

// Validate returns an error to tell whether the UpsertAwsRequest is valid or not.
func (r UpsertAwsRequest) Validate() error {
	return validator.New().Struct(r)
}

// IsValid returns a bool to tell whether the UpsertAwsRequest is valid or not.
func (r UpsertAwsRequest) IsValid() bool {
	return r.Validate() == nil
}
