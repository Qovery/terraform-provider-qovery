package annotations_group

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, organizationId string, request UpsertRequest) (*AnnotationsGroup, error)
	Get(ctx context.Context, organizationId string, annotationsGroupId string) (*AnnotationsGroup, error)
	Update(ctx context.Context, organizationId string, annotationsGroupId string, request UpsertRequest) (*AnnotationsGroup, error)
	Delete(ctx context.Context, organizationId string, annotationsGroupId string) error
}

type UpsertRequest struct {
	Name        string `validate:"required"`
	Annotations []AnnotationUpsertRequest
	Scopes      []string
}

type AnnotationUpsertRequest struct {
	Key   string `validate:"required"`
	Value string `validate:"required"`
}
