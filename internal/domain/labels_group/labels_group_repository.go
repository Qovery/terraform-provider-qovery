//go:generate mockery --testonly --with-expecter --name=Repository --structname=LabelsGroupRepository --filename=labels_group_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

package labels_group

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, organizationId string, request UpsertRequest) (*LabelsGroup, error)
	Get(ctx context.Context, organizationId string, labelsGroupId string) (*LabelsGroup, error)
	Update(ctx context.Context, organizationId string, labelsGroupId string, request UpsertRequest) (*LabelsGroup, error)
	Delete(ctx context.Context, organizationId string, labelsGroupId string) error
}

type UpsertRequest struct {
	Name   string `validate:"required"`
	Labels []LabelUpsertRequest
}

type LabelUpsertRequest struct {
	Key                      string `validate:"required"`
	Value                    string `validate:"required"`
	PropagateToCloudProvider bool   `validate:"required"`
}
