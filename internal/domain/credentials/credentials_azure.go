package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidUpsertAzureRequest is returned if an Azure Credentials upsert request is invalid.
	ErrInvalidUpsertAzureRequest = errors.New("invalid credentials upsert azure request")
)

// UpsertAzureRequest represents the parameters needed to create & update Azure Credentials.
type UpsertAzureRequest struct {
	Name                string `validate:"required"`
	AzureSubscriptionId string `validate:"required"`
	AzureTenantId       string `validate:"required"`
}

// Validate returns an error to tell whether the UpsertAzureRequest is valid or not.
func (r UpsertAzureRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertAzureRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertAzureRequest is valid or not.
func (r UpsertAzureRequest) IsValid() bool {
	return r.Validate() == nil
}

// AzureCredentials extends the base Credentials with Azure-specific fields returned by the API.
type AzureCredentials struct {
	Credentials
	AzureSubscriptionId      string
	AzureTenantId            string
	AzureApplicationId       string
	AzureApplicationObjectId string
}
