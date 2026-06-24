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
//
// natGatewaysCount: nil → nat_gateways key absent from features (simulates omitted block);
//
//	< 0  → nat_gateways is null;
//	>= 1 → nat_gateways object with static_ips_count = natGatewaysCount.
//
// natGatewaysEnabled: meaningful only when natGatewaysCount >= 1; sets static_ips_enabled.
// staticIP: nil → static_ip key absent; otherwise Bool value.
func makeValidateFeatures(natGatewaysCount *int64, natGatewaysEnabled bool, staticIP *bool) types.Object {
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
			"static_ips_enabled": types.BoolValue(natGatewaysEnabled),
			"static_ips_count":   types.Int64Value(*natGatewaysCount),
		})
		attrTypes[featureKeyNatGateways] = ngObjType
	}

	return types.ObjectValueMust(attrTypes, attrs)
}

func ptr[T any](v T) *T { return &v }

func TestValidateNatGatewaysConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		cloudProvider      types.String
		natGatewaysCount   *int64 // nil=key absent, <0=null, >=1=explicit count
		natGatewaysEnabled bool   // static_ips_enabled value (only used when count >= 1)
		staticIP           *bool  // nil=absent, otherwise value
		featuresNull       bool   // true → pass types.ObjectNull (simulates omitted features block)
		wantErrors         int
		wantWarnings       int
	}{
		{
			name:               "GCP+static_ip=true+enabled=true+count=3 → no diags",
			cloudProvider:      types.StringValue("GCP"),
			natGatewaysCount:   ptr(int64(3)),
			natGatewaysEnabled: true,
			staticIP:           ptr(true),
			wantErrors:         0,
			wantWarnings:       0,
		},
		{
			name:               "GCP+static_ip=true+enabled=true+count=1 → no diags",
			cloudProvider:      types.StringValue("GCP"),
			natGatewaysCount:   ptr(int64(1)),
			natGatewaysEnabled: true,
			staticIP:           ptr(true),
			wantErrors:         0,
			wantWarnings:       0,
		},
		{
			// Rule A error + Rule C warning
			name:               "AWS+enabled=true → Rule A error + Rule C warning",
			cloudProvider:      types.StringValue("AWS"),
			natGatewaysCount:   ptr(int64(1)),
			natGatewaysEnabled: true,
			staticIP:           ptr(true),
			wantErrors:         1,
			wantWarnings:       1,
		},
		{
			// Rule A error + Rule C warning (count > 1)
			name:               "AWS+count=3 → Rule A error + Rule C warning",
			cloudProvider:      types.StringValue("AWS"),
			natGatewaysCount:   ptr(int64(3)),
			natGatewaysEnabled: false,
			staticIP:           ptr(true),
			wantErrors:         1,
			wantWarnings:       1,
		},
		{
			// Rule B error: enabled=true but static_ip=false
			name:               "GCP+static_ip=false+enabled=true → Rule B error",
			cloudProvider:      types.StringValue("GCP"),
			natGatewaysCount:   ptr(int64(1)),
			natGatewaysEnabled: true,
			staticIP:           ptr(false),
			wantErrors:         1,
			wantWarnings:       0,
		},
		{
			// Rule B error: enabled=true with static_ip key entirely absent —
			// treated as the config-default false, same as an explicit false.
			name:               "GCP+static_ip absent+enabled=true → Rule B error",
			cloudProvider:      types.StringValue("GCP"),
			natGatewaysCount:   ptr(int64(1)),
			natGatewaysEnabled: true,
			staticIP:           nil,
			wantErrors:         1,
			wantWarnings:       0,
		},
		{
			// Rule B2 warning: count>1 but enabled=false
			name:               "GCP+static_ip=true+enabled=false+count=3 → Rule B2 warning",
			cloudProvider:      types.StringValue("GCP"),
			natGatewaysCount:   ptr(int64(3)),
			natGatewaysEnabled: false,
			staticIP:           ptr(true),
			wantErrors:         0,
			wantWarnings:       1,
		},
		{
			name:          "everything omitted (features null) → no diags",
			cloudProvider: types.StringValue("AWS"),
			featuresNull:  true,
			wantErrors:    0,
			wantWarnings:  0,
		},
		{
			// Rule C warning only: explicit default block {false,1} on non-GCP
			name:               "AWS+default {false,1} explicit → only Rule C warning",
			cloudProvider:      types.StringValue("AWS"),
			natGatewaysCount:   ptr(int64(1)),
			natGatewaysEnabled: false,
			staticIP:           ptr(false),
			wantErrors:         0,
			wantWarnings:       1,
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
			name:               "cloud_provider unknown → no diags (skip checks)",
			cloudProvider:      types.StringUnknown(),
			natGatewaysCount:   ptr(int64(3)),
			natGatewaysEnabled: false,
			staticIP:           ptr(false),
			wantErrors:         0,
			wantWarnings:       0,
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
				features = makeValidateFeatures(tc.natGatewaysCount, tc.natGatewaysEnabled, tc.staticIP)
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

// makeGkeKmsKeyFeatures builds a minimal features object containing only gke_kms_key.
func makeGkeKmsKeyFeatures(gkeKmsKey types.String) types.Object {
	attrs := map[string]attr.Value{featureKeyGkeKmsKey: gkeKmsKey}
	attrTypes := map[string]attr.Type{featureKeyGkeKmsKey: types.StringType}
	return types.ObjectValueMust(attrTypes, attrs)
}

func TestValidateGkeKmsKeyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cloudProvider types.String
		gkeKmsKey     types.String
		featuresNull  bool
		wantErrors    int
	}{
		{
			name:          "GCP + non-empty key → no error",
			cloudProvider: types.StringValue("GCP"),
			gkeKmsKey:     types.StringValue("projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-key"),
			wantErrors:    0,
		},
		{
			name:          "AWS + non-empty key → error",
			cloudProvider: types.StringValue("AWS"),
			gkeKmsKey:     types.StringValue("projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-key"),
			wantErrors:    1,
		},
		{
			name:          "SCW + non-empty key → error",
			cloudProvider: types.StringValue("SCW"),
			gkeKmsKey:     types.StringValue("projects/p/locations/l/keyRings/r/cryptoKeys/k"),
			wantErrors:    1,
		},
		{
			name:          "AWS + empty key → no error",
			cloudProvider: types.StringValue("AWS"),
			gkeKmsKey:     types.StringValue(""),
			wantErrors:    0,
		},
		{
			name:          "AWS + null key → no error",
			cloudProvider: types.StringValue("AWS"),
			gkeKmsKey:     types.StringNull(),
			wantErrors:    0,
		},
		{
			name:          "GCP + null key → no error",
			cloudProvider: types.StringValue("GCP"),
			gkeKmsKey:     types.StringNull(),
			wantErrors:    0,
		},
		{
			name:          "cloud_provider unknown + non-empty key → no error (benefit of the doubt)",
			cloudProvider: types.StringUnknown(),
			gkeKmsKey:     types.StringValue("projects/p/locations/l/keyRings/r/cryptoKeys/k"),
			wantErrors:    0,
		},
		{
			name:          "features null → no error",
			cloudProvider: types.StringValue("AWS"),
			featuresNull:  true,
			wantErrors:    0,
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
				features = makeGkeKmsKeyFeatures(tc.gkeKmsKey)
			}

			diags := validateGkeKmsKeyConfig(tc.cloudProvider, features)

			errCount := 0
			for _, d := range diags {
				if d.Severity() == diag.SeverityError {
					errCount++
				}
			}
			assert.Equal(t, tc.wantErrors, errCount, "error count mismatch")
		})
	}
}
