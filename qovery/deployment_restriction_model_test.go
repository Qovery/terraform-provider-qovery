//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	qovery "github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
)

func TestFromDeploymentRestrictionList(t *testing.T) {
	t.Parallel()

	t.Run("import_returns_api_restrictions_with_ids", func(t *testing.T) {
		t.Parallel()
		// On import the prior state has no restrictions (null set) but the API returns
		// existing ones. They must be reflected in state with their IDs so the next
		// apply reconciles by ID instead of trying to recreate them (which 409s).
		id := "restriction-id"
		apiRestrictions := []deploymentrestriction.ServiceDeploymentRestriction{
			{
				Id:    &id,
				Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH,
				Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
				Value: ".Dockerignore",
			},
		}

		result := FromDeploymentRestrictionList(types.SetNull(deploymentRestrictionObjectType), apiRestrictions)

		assert.False(t, result.IsNull(), "API restrictions must not be discarded on import")
		assert.Len(t, result.Elements(), 1)

		var list []DeploymentRestriction
		assert.False(t, result.ElementsAs(context.Background(), &list, false).HasError())
		if assert.Len(t, list, 1) {
			assert.Equal(t, id, list[0].Id.ValueString())
			assert.Equal(t, "MATCH", list[0].Mode.ValueString())
			assert.Equal(t, "PATH", list[0].Type.ValueString())
			assert.Equal(t, ".Dockerignore", list[0].Value.ValueString())
		}
	})

	t.Run("no_api_restrictions_and_null_state_returns_null", func(t *testing.T) {
		t.Parallel()
		result := FromDeploymentRestrictionList(types.SetNull(deploymentRestrictionObjectType), nil)

		assert.True(t, result.IsNull(), "a never-configured block should stay null")
	})

	t.Run("no_api_restrictions_and_empty_state_returns_empty", func(t *testing.T) {
		t.Parallel()
		emptyState := types.SetValueMust(deploymentRestrictionObjectType, []attr.Value{})

		result := FromDeploymentRestrictionList(emptyState, nil)

		assert.False(t, result.IsNull(), "an explicitly empty set should stay empty, not null")
		assert.Len(t, result.Elements(), 0)
	})
}
