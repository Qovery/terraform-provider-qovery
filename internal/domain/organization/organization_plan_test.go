package organization_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// TestNewPlanFromString validate that the plans qovery.PlanEnum defined in Qovery's API Client are valid.
// This is useful to make sure the organization.Plan stays up to date.
func TestNewPlanFromString(t *testing.T) {
	t.Parallel()

	for _, qoveryPlan := range qovery.AllowedPlanEnumEnumValues {
		qoveryPlanStr := string(qoveryPlan)
		t.Run(qoveryPlanStr, func(t *testing.T) {
			plan, err := organization.NewPlanFromString(qoveryPlanStr)
			assert.NoError(t, err)
			assert.Equal(t, plan.String(), qoveryPlanStr)
		})
	}
}
