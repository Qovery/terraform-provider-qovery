//go:build unit && !integration
// +build unit,!integration

package labels_group_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
)

func TestLabelsGroup_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name        string
		group       labels_group.LabelsGroup
		expectError bool
	}{
		{
			name: "valid labels group",
			group: labels_group.LabelsGroup{
				Id:   uuid.New(),
				Name: "test-group",
				Labels: []qovery.Label{
					{Key: "key1", Value: "value1", PropagateToCloudProvider: true},
				},
			},
			expectError: false,
		},
		{
			name: "valid labels group with empty labels",
			group: labels_group.LabelsGroup{
				Id:     uuid.New(),
				Name:   "test-group",
				Labels: []qovery.Label{},
			},
			expectError: false,
		},
		{
			name: "valid labels group with nil labels",
			group: labels_group.LabelsGroup{
				Id:   uuid.New(),
				Name: "test-group",
			},
			expectError: false,
		},
		{
			name: "invalid labels group with zero id",
			group: labels_group.LabelsGroup{
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
		request     labels_group.UpsertRequest
		expectError bool
	}{
		{
			name: "valid upsert request",
			request: labels_group.UpsertRequest{
				Name: "test-group",
				Labels: []labels_group.LabelUpsertRequest{
					{Key: "key1", Value: "value1", PropagateToCloudProvider: true},
				},
			},
			expectError: false,
		},
		{
			name: "valid upsert request with empty labels",
			request: labels_group.UpsertRequest{
				Name:   "test-group",
				Labels: []labels_group.LabelUpsertRequest{},
			},
			expectError: false,
		},
		{
			name: "invalid upsert request with empty name",
			request: labels_group.UpsertRequest{
				Name:   "",
				Labels: []labels_group.LabelUpsertRequest{},
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

func TestLabelUpsertRequest_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name        string
		request     labels_group.LabelUpsertRequest
		expectError bool
	}{
		{
			name: "valid label upsert request",
			request: labels_group.LabelUpsertRequest{
				Key:                      "key1",
				Value:                    "value1",
				PropagateToCloudProvider: true,
			},
			expectError: false,
		},
		{
			name: "invalid label upsert request with propagate false",
			request: labels_group.LabelUpsertRequest{
				Key:                      "key1",
				Value:                    "value1",
				PropagateToCloudProvider: false,
			},
			expectError: true, // required tag means false (zero value) is invalid
		},
		{
			name: "invalid label upsert request with empty key",
			request: labels_group.LabelUpsertRequest{
				Key:                      "",
				Value:                    "value1",
				PropagateToCloudProvider: true,
			},
			expectError: true,
		},
		{
			name: "invalid label upsert request with empty value",
			request: labels_group.LabelUpsertRequest{
				Key:                      "key1",
				Value:                    "",
				PropagateToCloudProvider: true,
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
	request := labels_group.UpsertServiceRequest{
		LabelsGroupUpsertRequest: labels_group.UpsertRequest{
			Name: "test-group",
		},
	}

	err := request.Validate()
	assert.NoError(t, err)
}
