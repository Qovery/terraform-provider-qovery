package organization

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilOrganization is the error return if an Organization is nil.
	ErrNilOrganization = errors.New("organization cannot be nil")
	// ErrInvalidOrganization is the error return if an Organization is invalid.
	ErrInvalidOrganization = errors.New("invalid organization")
	// ErrInvalidOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidNameParam is returned if the name param is invalid.
	ErrInvalidNameParam = errors.New("invalid name param")
	// ErrInvalidPlanParam is returned if the plan param is invalid.
	ErrInvalidPlanParam = errors.New("invalid plan param")
	// ErrInvalidUpdateRequest is returned if the organization update request is invalid.
	ErrInvalidUpdateRequest = errors.New("invalid organization update request")
)

// Organization represents the domain model for a Qovery organization.
type Organization struct {
	ID          uuid.UUID `validate:"required"`
	Name        string    `validate:"required"`
	Plan        Plan      `validate:"required"`
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

// NewOrganizationParams represents the arguments needed to create an Organization.
type NewOrganizationParams struct {
	OrganizationID string
	Name           string
	Plan           string
	Description    *string
}

// NewOrganization returns a new instance of an Organization domain model.
func NewOrganization(params NewOrganizationParams) (*Organization, error) {
	organizationUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidOrganizationIDParam.Error())
	}

	plan, err := NewPlanFromString(params.Plan)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidPlanParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidNameParam
	}

	orga := &Organization{
		ID:          organizationUUID,
		Name:        params.Name,
		Plan:        *plan,
		Description: params.Description,
	}

	if err := orga.Validate(); err != nil {
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
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpdateRequest domain request is valid or not.
func (r UpdateRequest) IsValid() bool {
	return r.Validate() == nil
}
