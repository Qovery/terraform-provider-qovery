package organization

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	// ErrNilOrganization is the error return if an Organization is nil.
	ErrNilOrganization = errors.New("organization cannot be nil")
	// ErrInvalidOrganization is the error return if an Organization is invalid.
	ErrInvalidOrganization = errors.New("invalid organization")
)

// Organization represents the domain model for a Qovery organization.
type Organization struct {
	ID          string `validate:"required"`
	Name        string `validate:"required"`
	Plan        Plan   `validate:"required"`
	Description *string
}

// Validate returns an error to tell whether the Organization domain model is valid or not.
func (o Organization) Validate() error {
	return validator.New().Struct(o)
}

// IsValid returns a bool to tell whether the Organization domain model is valid or not.
func (o Organization) IsValid() bool {
	return o.Validate() == nil
}

// NewOrganization returns a new instance of an Organization domain model.
func NewOrganization(id string, name string, plan Plan) (*Organization, error) {
	orga := &Organization{
		ID:   id,
		Name: name,
		Plan: plan,
	}

	if err := orga.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidOrganization.Error())
	}

	if err := orga.Plan.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidOrganization.Error())
	}

	return orga, nil
}

// UpdateRequest represents the parameters needed to update an Organization.
type UpdateRequest struct {
	Name        string `validate:"required"`
	Description *string
}

// Validate returns an error to tell whether the UpdateRequest domain request is valid or not.
func (r UpdateRequest) Validate() error {
	return validator.New().Struct(r)
}

// IsValid returns a bool to tell whether the UpdateRequest domain request is valid or not.
func (r UpdateRequest) IsValid() bool {
	return r.Validate() == nil
}
