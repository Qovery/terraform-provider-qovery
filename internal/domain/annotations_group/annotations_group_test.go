//go:build unit && !integration
// +build unit,!integration

package annotations_group_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
)

func TestAnnotationsGroup_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name        string
		group       annotations_group.AnnotationsGroup
		expectError bool
	}{
		{
			name: "valid annotations group",
			group: annotations_group.AnnotationsGroup{
				Id:   uuid.New(),
				Name: "test-group",
				Annotations: []qovery.Annotation{
					{Key: "key1", Value: "value1"},
				},
				Scopes: []qovery.OrganizationAnnotationsGroupScopeEnum{
					qovery.ORGANIZATIONANNOTATIONSGROUPSCOPEENUM_PODS,
				},
			},
			expectError: false,
		},
		{
			name: "valid annotations group with empty annotations",
			group: annotations_group.AnnotationsGroup{
				Id:          uuid.New(),
				Name:        "test-group",
				Annotations: []qovery.Annotation{},
				Scopes:      []qovery.OrganizationAnnotationsGroupScopeEnum{},
			},
			expectError: false,
		},
		{
			name: "valid annotations group with nil annotations",
			group: annotations_group.AnnotationsGroup{
				Id:   uuid.New(),
				Name: "test-group",
			},
			expectError: false,
		},
		{
			name: "invalid annotations group with zero id",
			group: annotations_group.AnnotationsGroup{
				Id:   uuid.UUID{},
				Name: "test-group",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.group)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpsertRequest_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name        string
		request     annotations_group.UpsertRequest
		expectError bool
	}{
		{
			name: "valid upsert request",
			request: annotations_group.UpsertRequest{
				Name: "test-group",
				Annotations: []annotations_group.AnnotationUpsertRequest{
					{Key: "key1", Value: "value1"},
				},
				Scopes: []string{"PODS"},
			},
			expectError: false,
		},
		{
			name: "valid upsert request with empty annotations",
			request: annotations_group.UpsertRequest{
				Name:        "test-group",
				Annotations: []annotations_group.AnnotationUpsertRequest{},
				Scopes:      []string{},
			},
			expectError: false,
		},
		{
			name: "invalid upsert request with empty name",
			request: annotations_group.UpsertRequest{
				Name:        "",
				Annotations: []annotations_group.AnnotationUpsertRequest{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAnnotationUpsertRequest_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name        string
		request     annotations_group.AnnotationUpsertRequest
		expectError bool
	}{
		{
			name: "valid annotation upsert request",
			request: annotations_group.AnnotationUpsertRequest{
				Key:   "key1",
				Value: "value1",
			},
			expectError: false,
		},
		{
			name: "invalid annotation upsert request with empty key",
			request: annotations_group.AnnotationUpsertRequest{
				Key:   "",
				Value: "value1",
			},
			expectError: true,
		},
		{
			name: "invalid annotation upsert request with empty value",
			request: annotations_group.AnnotationUpsertRequest{
				Key:   "key1",
				Value: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpsertServiceRequest_Validate(t *testing.T) {
	t.Parallel()

	// UpsertServiceRequest.Validate() currently returns nil
	// This test documents the current behavior
	request := annotations_group.UpsertServiceRequest{
		AnnotationsGroupUpsertRequest: annotations_group.UpsertRequest{
			Name: "test-group",
		},
	}

	err := request.Validate()
	assert.NoError(t, err)
}
