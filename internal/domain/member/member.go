package member

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidMember              = errors.New("invalid organization member")
	ErrInvalidOrganizationIdParam = errors.New("invalid organization id param")
	ErrInvalidEmailParam          = errors.New("invalid email param")
	ErrInvalidInviteRequest       = errors.New("invalid organization member invite request")
	ErrInvalidUpdateRoleRequest   = errors.New("invalid organization member update role request")
)

// Invitation statuses exposed in the terraform state. PENDING and EXPIRED come from the
// invitation API; ACCEPTED is synthesized when the email is found in the active member list.
const (
	StatusPending  = "PENDING"
	StatusExpired  = "EXPIRED"
	StatusAccepted = "ACCEPTED"
)

// Member represents an organization member in either lifecycle state. While the invitation
// is pending, ID is the invite UUID and UserID is nil; once accepted, ID and UserID both
// hold the user id (the identity provider subject, e.g. `github|12345` — not a UUID).
type Member struct {
	ID               string    `validate:"required"`
	OrganizationID   uuid.UUID `validate:"required"`
	Email            string    `validate:"required,email"`
	RoleID           *string
	UserID           *string
	InvitationStatus string `validate:"required"`
}

func (m Member) Validate() error {
	if err := validator.New().Struct(m); err != nil {
		return errors.Wrap(err, ErrInvalidMember.Error())
	}
	return nil
}

type InviteRequest struct {
	Email  string `validate:"required,email"`
	RoleID string `validate:"required,uuid"`
}

func (r InviteRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidInviteRequest.Error())
	}
	return nil
}

type UpdateRoleRequest struct {
	RoleID string `validate:"required,uuid"`
}

func (r UpdateRoleRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRoleRequest.Error())
	}
	return nil
}

// ValidateEmail checks a standalone email parameter (Get/Update/Delete key).
func ValidateEmail(email string) error {
	if err := validator.New().Var(email, "required,email"); err != nil {
		return errors.Wrap(err, ErrInvalidEmailParam.Error())
	}
	return nil
}
