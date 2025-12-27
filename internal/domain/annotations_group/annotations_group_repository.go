//go:generate mockery --testonly --with-expecter --name=Repository --structname=AnnotationsGroupRepository --filename=annotations_group_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

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
