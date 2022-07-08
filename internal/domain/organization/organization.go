package organization

import (
	"context"
)

type API interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}

type Service interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}

type Organization struct {
	ID          string
	Name        string
	Plan        Plan
	Description *string
}

func (o Organization) WithDescription(description *string) Organization {
	o.Description = description
	return o
}

func NewOrganization(id string, name string, plan Plan) *Organization {
	return &Organization{
		ID:   id,
		Name: name,
		Plan: plan,
	}
}

type UpdateRequest struct {
	Name        string
	Description *string
}
