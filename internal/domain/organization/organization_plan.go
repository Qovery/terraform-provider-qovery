package organization

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Plan string

const (
	PLAN_FREE         Plan = "FREE"
	PLAN_PROFESSIONAL Plan = "PROFESSIONAL"
	PLAN_BUSINESS     Plan = "BUSINESS"
	PLAN_ENTERPRISE   Plan = "ENTERPRISE"
)

var AllowedPlanValues = []Plan{
	"FREE",
	"PROFESSIONAL",
	"BUSINESS",
	"ENTERPRISE",
}

func NewPlanFromString(v string) (*Plan, error) {
	ev := Plan(v)
	if ev.IsValid() {
		return &ev, nil
	}
	return nil, fmt.Errorf("invalid value '%v' for Plan: valid values are %v", v, AllowedPlanValues)
}

func (v Plan) String() string {
	return string(v)
}

func (v Plan) IsValid() bool {
	return slices.Contains(AllowedPlanValues, v)
}
