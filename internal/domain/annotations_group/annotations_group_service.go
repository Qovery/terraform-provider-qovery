package annotations_group

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, organizationId string, request UpsertServiceRequest) (*AnnotationsGroup, error)
	Get(ctx context.Context, organizationId string, annotationsGroupID string) (*AnnotationsGroup, error)
	Update(ctx context.Context, organizationId string, annotationsGroupID string, request UpsertServiceRequest) (*AnnotationsGroup, error)
	Delete(ctx context.Context, organizationId string, annotationsGroupID string) error
}

type UpsertServiceRequest struct {
	AnnotationsGroupUpsertRequest UpsertRequest
}

func (r UpsertServiceRequest) Validate() error {

	return nil
}
