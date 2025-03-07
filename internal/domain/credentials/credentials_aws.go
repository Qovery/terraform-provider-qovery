package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrInvalidUpsertAwsRequest = errors.New("invalid credentials upsert aws request")
)

type AwsStaticCredentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}
type AwsRoleCredentials struct {
	RoleArn string `validate:"required"`
}

// UpsertAwsRequest represents the parameters needed to create & update AWS Credentials.
type UpsertAwsRequest struct {
	Name              string `validate:"required"`
	StaticCredentials *AwsStaticCredentials
	RoleCredentials   *AwsRoleCredentials
}

// Validate returns an error to tell whether the UpsertAwsRequest is valid or not.
func (r UpsertAwsRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertAwsRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertAwsRequest is valid or not.
func (r UpsertAwsRequest) IsValid() bool {
	return r.Validate() == nil
}
