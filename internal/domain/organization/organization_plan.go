package organization

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// Plan is an enum that contains all the valid values of an organization plan.
type Plan string

const (
	PlanFree             Plan = "FREE"
	PlanTeam             Plan = "TEAM"
	PlanTeamYearly       Plan = "TEAM_YEARLY"
	PlanEnterprise       Plan = "ENTERPRISE"
	PlanEnterpriseYearly Plan = "ENTERPRISE_YEARLY"
	PlanProfessional     Plan = "PROFESSIONAL"
	PlanBusiness         Plan = "BUSINESS"
	PlanUser2025         Plan = "USER_2025"
	PlanTeam2025         Plan = "TEAM_2025"
	PlanBusiness2025     Plan = "BUSINESS_2025"
	PlanEnterprise2025   Plan = "ENTERPRISE_2025"
)

// AllowedPlanValues contains all the valid values of a Plan.
var AllowedPlanValues = []Plan{
	PlanFree,
	PlanTeam,
	PlanTeamYearly,
	PlanEnterprise,
	PlanEnterpriseYearly,
	PlanProfessional,
	PlanBusiness,
	PlanUser2025,
	PlanTeam2025,
	PlanBusiness2025,
	PlanEnterprise2025,
}

// String returns the string value of a Plan.
func (v Plan) String() string {
	return string(v)
}

// Validate returns an error to tell whether the Plan is valid or not.
func (v Plan) Validate() error {
	if slices.Contains(AllowedPlanValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Plan: valid values are %v", v, AllowedPlanValues)
}

// IsValid returns a bool to tell whether the Plan is valid or not.
func (v Plan) IsValid() bool {
	return v.Validate() == nil
}

// NewPlanFromString tries to turn a string into a Plan.
// It returns an error if the string is not a valid value.
func NewPlanFromString(v string) (*Plan, error) {
	ev := Plan(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
