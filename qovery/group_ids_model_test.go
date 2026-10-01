//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
)

func stringSetValue(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elements)
}

// TestGroupIdsFromAPI covers the labels_group_ids / annotations_group_ids read of every service:
// the API value always wins, and an empty API value keeps the prior shape (null or []).
func TestGroupIdsFromAPI(t *testing.T) {
	t.Parallel()

	converters := map[string]func(prior types.Set, ids []string) types.Set{
		// container and job
		"string_list": stringSetFromAPI,
		// application and database
		"labels_group_response_list": func(prior types.Set, ids []string) types.Set {
			groups := make([]qovery.OrganizationLabelsGroupResponse, 0, len(ids))
			for _, id := range ids {
				groups = append(groups, qovery.OrganizationLabelsGroupResponse{Id: id})
			}
			return fromLabelsGroupResponseList(prior, groups)
		},
		"annotations_group_response_list": func(prior types.Set, ids []string) types.Set {
			groups := make([]qovery.OrganizationAnnotationsGroupResponse, 0, len(ids))
			for _, id := range ids {
				groups = append(groups, qovery.OrganizationAnnotationsGroupResponse{Id: id})
			}
			return fromAnnotationsGroupResponseList(prior, groups)
		},
	}

	testCases := []struct {
		name  string
		prior types.Set
		api   []string
		want  types.Set
	}{
		{
			// Import, or groups attached in the Console while the config omits the attribute.
			name:  "prior_null_with_api_values_reports_api_values",
			prior: types.SetNull(types.StringType),
			api:   []string{"group-1", "group-2"},
			want:  stringSetValue("group-1", "group-2"),
		},
		{
			name:  "prior_null_with_empty_api_value_stays_null",
			prior: types.SetNull(types.StringType),
			api:   nil,
			want:  types.SetNull(types.StringType),
		},
		{
			name:  "prior_empty_with_empty_api_value_stays_empty",
			prior: stringSetValue(),
			api:   []string{},
			want:  stringSetValue(),
		},
		{
			name:  "prior_set_reports_api_values",
			prior: stringSetValue("group-1"),
			api:   []string{"group-2"},
			want:  stringSetValue("group-2"),
		},
		{
			// Groups detached in the Console while the config declares them.
			name:  "prior_set_with_empty_api_value_reports_empty",
			prior: stringSetValue("group-1"),
			api:   nil,
			want:  stringSetValue(),
		},
	}

	for converterName, convert := range converters {
		convert := convert
		for _, tc := range testCases {
			tc := tc
			t.Run(converterName+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, convert(tc.prior, tc.api))
			})
		}
	}
}
