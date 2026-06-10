//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// makeValidateFeatures builds a features types.Object suitable for validateNatGatewaysConfig.
// natGatewaysCount: nil → nat_gateways key absent from features (simulates omitted block);
//
//	< 0  → nat_gateways is null;
//	>= 1 → nat_gateways object with static_ips_count = natGatewaysCount.
//
// staticIP: nil → static_ip key absent; otherwise Bool value.
func makeValidateFeatures(natGatewaysCount *int64, staticIP *bool) types.Object {
	attrs := map[string]attr.Value{}
	attrTypes := map[string]attr.Type{}

	// static_ip
	if staticIP != nil {
		attrs[featureKeyStaticIP] = types.BoolValue(*staticIP)
		attrTypes[featureKeyStaticIP] = types.BoolType
	}

	// nat_gateways
	ngAttrTypes := createNatGatewaysFeatureAttrTypes()
	ngObjType := types.ObjectType{AttrTypes: ngAttrTypes}
	if natGatewaysCount == nil {
		// key absent — don't add
	} else if *natGatewaysCount < 0 {
		attrs[featureKeyNatGateways] = types.ObjectNull(ngAttrTypes)
		attrTypes[featureKeyNatGateways] = ngObjType
	} else {
		attrs[featureKeyNatGateways] = types.ObjectValueMust(ngAttrTypes, map[string]attr.Value{
			"static_ips_count": types.Int64Value(*natGatewaysCount),
		})
		attrTypes[featureKeyNatGateways] = ngObjType
	}

	return types.ObjectValueMust(attrTypes, attrs)
}

func ptr[T any](v T) *T { return &v }

func TestValidateNatGatewaysConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		cloudProvider    types.String
		natGatewaysCount *int64 // nil=key absent, <0=null, >=1=explicit count
		staticIP         *bool  // nil=absent, otherwise value
		featuresNull     bool   // true → pass types.ObjectNull (simulates omitted features block)
		wantErrors       int
		wantWarnings     int
	}{
		{
			name:             "GCP+static_ip=true+count=3 → no diags",
			cloudProvider:    types.StringValue("GCP"),
			natGatewaysCount: ptr(int64(3)),
			staticIP:         ptr(true),
			wantErrors:       0,
			wantWarnings:     0,
		},
		{
			name:             "GCP+static_ip=true+count=1 → no diags",
			cloudProvider:    types.StringValue("GCP"),
			natGatewaysCount: ptr(int64(1)),
			staticIP:         ptr(true),
			wantErrors:       0,
			wantWarnings:     0,
		},
		{
			name:             "AWS+count=3 → error rule A",
			cloudProvider:    types.StringValue("AWS"),
			natGatewaysCount: ptr(int64(3)),
			staticIP:         ptr(true),
			wantErrors:       1,
			wantWarnings:     1, // Rule C warning fires before Rule A error
		},
		{
			name:             "GCP+static_ip=false+count=2 → error rule B",
			cloudProvider:    types.StringValue("GCP"),
			natGatewaysCount: ptr(int64(2)),
			staticIP:         ptr(false),
			wantErrors:       1,
			wantWarnings:     0,
		},
		{
			name:          "everything omitted (features null) → no diags",
			cloudProvider: types.StringValue("AWS"),
			featuresNull:  true,
			wantErrors:    0,
			wantWarnings:  0,
		},
		{
			name:             "AWS+count=1 explicit → warning only (Rule C)",
			cloudProvider:    types.StringValue("AWS"),
			natGatewaysCount: ptr(int64(1)),
			staticIP:         ptr(false),
			wantErrors:       0,
			wantWarnings:     1,
		},
		{
			name:             "nat_gateways null (omitted in config) → no diags",
			cloudProvider:    types.StringValue("GCP"),
			natGatewaysCount: ptr(int64(-1)), // null
			staticIP:         ptr(true),
			wantErrors:       0,
			wantWarnings:     0,
		},
		{
			name:             "cloud_provider unknown → no diags (skip checks)",
			cloudProvider:    types.StringUnknown(),
			natGatewaysCount: ptr(int64(3)),
			staticIP:         ptr(false),
			wantErrors:       0,
			wantWarnings:     0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var features types.Object
			if tc.featuresNull {
				features = types.ObjectNull(map[string]attr.Type{})
			} else {
				features = makeValidateFeatures(tc.natGatewaysCount, tc.staticIP)
			}

			diags := validateNatGatewaysConfig(tc.cloudProvider, features)

			errCount := 0
			warnCount := 0
			for _, d := range diags {
				switch d.Severity() {
				case diag.SeverityError:
					errCount++
				case diag.SeverityWarning:
					warnCount++
				}
			}

			assert.Equal(t, tc.wantErrors, errCount, "error count mismatch")
			assert.Equal(t, tc.wantWarnings, warnCount, "warning count mismatch")
		})
	}
}
