package labels_group

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, organizationId string, request UpsertServiceRequest) (*LabelsGroup, error)
	Get(ctx context.Context, organizationId string, labelsGroupID string) (*LabelsGroup, error)
	Update(ctx context.Context, organizationId string, labelsGroupID string, request UpsertServiceRequest) (*LabelsGroup, error)
	Delete(ctx context.Context, organizationId string, labelsGroupID string) error
}

type UpsertServiceRequest struct {
	LabelsGroupUpsertRequest UpsertRequest
}

func (r UpsertServiceRequest) Validate() error {

	return nil
}
