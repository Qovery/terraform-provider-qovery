package organization_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

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
