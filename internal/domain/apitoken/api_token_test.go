//go:build unit && !integration

package apitoken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

func TestCreateRequestValidate(t *testing.T) {
	t.Parallel()

	description := "my description"

	testCases := []struct {
		TestName      string
		Request       apitoken.CreateRequest
		ErrorContains string
	}{
		{
			TestName: "fail_with_empty_name",
			Request: apitoken.CreateRequest{
				Name:   "",
				RoleID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
			ErrorContains: apitoken.ErrInvalidCreateRequest.Error(),
		},
		{
			TestName: "fail_with_empty_role_id",
			Request: apitoken.CreateRequest{
				Name:   "my-token",
				RoleID: "",
			},
			ErrorContains: apitoken.ErrInvalidCreateRequest.Error(),
		},
		{
			TestName: "success_without_description",
			Request: apitoken.CreateRequest{
				Name:   "my-token",
				RoleID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
		},
		{
			TestName: "success_with_description",
			Request: apitoken.CreateRequest{
				Name:        "my-token",
				Description: &description,
				RoleID:      "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			err := tc.Request.Validate()
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				return
			}
			assert.NoError(t, err)
		})
	}
}
