package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

// Ensure organizationMemberService defined type fully satisfy the member.Service interface.
var _ member.Service = organizationMemberService{}

// organizationMemberService implements the interface member.Service.
type organizationMemberService struct {
	repo member.Repository
}

func NewOrganizationMemberService(repo member.Repository) (member.Service, error) {
	if repo == nil {
		return nil, ErrInvalidRepository
	}
	return &organizationMemberService{repo: repo}, nil
}

func (s organizationMemberService) Create(ctx context.Context, organizationID string, request member.InviteRequest) (*member.Member, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToInviteMember.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToInviteMember.Error())
	}
	res, err := s.repo.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToInviteMember.Error())
	}
	return res, nil
}

func (s organizationMemberService) Get(ctx context.Context, organizationID string, email string) (*member.Member, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToGetMember.Error())
	}
	if err := member.ValidateEmail(email); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToGetMember.Error())
	}
	res, err := s.repo.Get(ctx, organizationID, email)
	if err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToGetMember.Error())
	}
	return res, nil
}

func (s organizationMemberService) Update(ctx context.Context, organizationID string, email string, request member.UpdateRoleRequest) (*member.Member, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToUpdateMember.Error())
	}
	if err := member.ValidateEmail(email); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToUpdateMember.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToUpdateMember.Error())
	}
	res, err := s.repo.Update(ctx, organizationID, email, request)
	if err != nil {
		return nil, errors.Wrap(err, member.ErrFailedToUpdateMember.Error())
	}
	return res, nil
}

func (s organizationMemberService) Delete(ctx context.Context, organizationID string, email string) error {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, member.ErrFailedToDeleteMember.Error())
	}
	if err := member.ValidateEmail(email); err != nil {
		return errors.Wrap(err, member.ErrFailedToDeleteMember.Error())
	}
	if err := s.repo.Delete(ctx, organizationID, email); err != nil {
		return errors.Wrap(err, member.ErrFailedToDeleteMember.Error())
	}
	return nil
}

func (s organizationMemberService) checkOrganizationID(organizationID string) error {
	return validateUUIDParam(organizationID, member.ErrInvalidOrganizationIdParam)
}
