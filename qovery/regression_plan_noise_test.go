//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Asserts that every Computed attribute on every resource has a state-preserving
// plan modifier, a Default, or an explicit allowlist entry — preventing false
// "(known after apply)" flicker that cascades into "will be read during apply"
// noise on dependent data sources. Custom state-preserving modifiers must be
// added to preservesState() to be recognised.

// The framework's UseStateForUnknown modifiers have unexported concrete types,
// so we identify them by their Description() string — the only stable signal.
const useStateForUnknownDescription = "Once set, the value of this attribute in state will not change."

func preservesState(m planmodifier.Describer) bool {
	switch m.(type) {
	case useStateUnlessNameChangesModifier,
		useStateUnlessPortsChangeModifier,
		smartAllowApiOverrideModifier:
		return true
	}
	return m.Description(context.Background()) == useStateForUnknownDescription
}

func anyPreservesState[M planmodifier.Describer](mods []M) bool {
	for _, m := range mods {
		if preservesState(m) {
			return true
		}
	}
	return false
}

// flickerAllowlist tolerates known flickering attributes. Format:
// "<terraform_resource_type>.<dot.separated.attribute.path>" -> reason.
// Entries prefixed "TODO:" are pre-existing gaps to fix incrementally; once
// the attribute gets a state-preserving modifier, remove the entry. Entries
// without "TODO:" document legitimately-volatile attributes — their reason
// should explain why the flicker is correct.
var flickerAllowlist = map[string]string{
	"qovery_container_registry.description":             "TODO: add UseStateForUnknown",
	"qovery_database.external_host":                     "TODO: add UseStateForUnknown (consider UseStateUnlessAccessibilityChanges custom modifier)",
	"qovery_database.icon_uri":                          "TODO: add UseStateForUnknown",
	"qovery_database.instance_type":                     "TODO: add UseStateForUnknown",
	"qovery_database.internal_host":                     "TODO: add UseStateForUnknown",
	"qovery_database.login":                             "TODO: add UseStateForUnknown",
	"qovery_database.password":                          "TODO: add UseStateForUnknown",
	"qovery_database.port":                              "TODO: add UseStateForUnknown",
	"qovery_deployment.id":                              "TODO: add UseStateForUnknown (resource's own id flickers — strong bug)",
	"qovery_environment.built_in_environment_variables": "TODO: add UseStateUnlessNameChanges (top-level list; same pattern as service resources)",
	"qovery_git_token.bitbucket_workspace":              "TODO: add UseStateForUnknown",
	"qovery_git_token.description":                      "TODO: add UseStateForUnknown",
	"qovery_helm_repository.description":                "TODO: add UseStateForUnknown",
	"qovery_organization.description":                   "TODO: add UseStateForUnknown",
	"qovery_project.built_in_environment_variables":     "TODO: add UseStateUnlessNameChanges (top-level list; same pattern as service resources)",
	"qovery_project.description":                        "TODO: add UseStateForUnknown",
	"qovery_scaleway_credentials.id":                    "TODO: add UseStateForUnknown (resource's own id flickers — strong bug)",
	"qovery_terraform_service.advanced_settings_json":   "TODO: add UseStateForUnknown",

	"qovery_application.git_repository.branch":                       "TODO: add UseStateForUnknown",
	"qovery_helm.source.git_repository.branch":                       "TODO: add UseStateForUnknown",
	"qovery_helm.source.git_repository.git_token_id":                 "TODO: add UseStateForUnknown",
	"qovery_helm.values_override.file.git_repository.git_token_id":   "TODO: add UseStateForUnknown",
	"qovery_job.schedule.cronjob.command.entrypoint":                 "TODO: add UseStateForUnknown",
	"qovery_job.schedule.on_delete.entrypoint":                       "TODO: add UseStateForUnknown",
	"qovery_job.schedule.on_start.entrypoint":                        "TODO: add UseStateForUnknown",
	"qovery_job.schedule.on_stop.entrypoint":                         "TODO: add UseStateForUnknown",
	"qovery_job.source.docker.git_repository.root_path":              "TODO: add UseStateForUnknown",

	"qovery_application.built_in_environment_variables.description":  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.built_in_environment_variables.id":           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.built_in_environment_variables.key":          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.built_in_environment_variables.value":        "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.custom_domains.id":                           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.custom_domains.status":                       "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.custom_domains.validation_domain":            "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.deployment_restrictions.id":                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.environment_variable_aliases.id":             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.environment_variable_files.id":               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.environment_variable_overrides.id":           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.environment_variables.id":                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.secret_aliases.id":                           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.secret_files.id":                             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.secret_overrides.id":                         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.secrets.id":                                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_application.storage.id":                                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.built_in_environment_variables.description":    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.built_in_environment_variables.id":             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.built_in_environment_variables.key":            "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.built_in_environment_variables.value":          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.custom_domains.id":                             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.custom_domains.status":                         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.custom_domains.validation_domain":              "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.environment_variable_aliases.id":               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.environment_variable_files.id":                 "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.environment_variable_overrides.id":             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.environment_variables.id":                      "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.ports.protocol":                                "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.secret_aliases.id":                             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.secret_files.id":                               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.secret_overrides.id":                           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.secrets.id":                                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_container.storage.id":                                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.built_in_environment_variables.description":  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.built_in_environment_variables.id":           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.built_in_environment_variables.key":          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.built_in_environment_variables.value":        "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.environment_variable_aliases.id":             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.environment_variable_files.id":               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.environment_variable_overrides.id":           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.environment_variables.id":                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.secret_aliases.id":                           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.secret_files.id":                             "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.secret_overrides.id":                         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_environment.secrets.id":                                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.built_in_environment_variables.description":         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.built_in_environment_variables.id":                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.built_in_environment_variables.key":                 "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.built_in_environment_variables.value":               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.custom_domains.id":                                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.custom_domains.status":                              "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.custom_domains.validation_domain":                   "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.deployment_restrictions.id":                         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.environment_variable_aliases.id":                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.environment_variable_files.id":                      "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.environment_variable_overrides.id":                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.environment_variables.id":                           "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.ports.protocol":                                     "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.secret_aliases.id":                                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.secret_files.id":                                    "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.secret_overrides.id":                                "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_helm.secrets.id":                                         "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.built_in_environment_variables.description":          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.built_in_environment_variables.id":                   "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.built_in_environment_variables.key":                  "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.built_in_environment_variables.value":                "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.deployment_restrictions.id":                          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.environment_variable_aliases.id":                     "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.environment_variable_files.id":                       "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.environment_variable_overrides.id":                   "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.environment_variables.id":                            "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.secret_aliases.id":                                   "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.secret_files.id":                                     "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.secret_overrides.id":                                 "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_job.secrets.id":                                          "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.built_in_environment_variables.description":      "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.built_in_environment_variables.id":               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.built_in_environment_variables.key":              "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.built_in_environment_variables.value":            "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.environment_variable_aliases.id":                 "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.environment_variable_files.id":                   "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.environment_variables.id":                        "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.secret_aliases.id":                               "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.secret_files.id":                                 "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",
	"qovery_project.secrets.id":                                      "TODO: add UseStateForUnknown (set-element id; cosmetic flicker only)",

	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_a_ids":     "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_b_ids":     "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_c_ids":     "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.eks_create_nodes_in_private_subnet": "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_a_ids":    "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_b_ids":    "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_c_ids":    "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_a_ids":            "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_b_ids":            "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_c_ids":            "TODO: legitimate volatility? — verify with cluster owner; if VPC swap recomputes, document; else add UseStateForUnknown",
	"qovery_terraform_service.created_at":                                    "TODO: legitimate volatility? — if API restamps on every write, replace with permanent reason (no TODO prefix); else add UseStateForUnknown",
	"qovery_terraform_service.updated_at":                                    "TODO: legitimate volatility? — if API restamps on every write, replace with permanent reason (no TODO prefix); else add UseStateForUnknown",
}

type attributeStatus struct {
	computed       bool
	preservesState bool
	hasDefault     bool
}

// inspectAttribute returns ok=false for unrecognised attribute types so the
// caller can fail loudly rather than silently mis-classify them.
func inspectAttribute(attr schema.Attribute) (status attributeStatus, ok bool) {
	switch a := attr.(type) {
	case schema.StringAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.BoolAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.Int64Attribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.ListAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SetAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.MapAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SingleNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.ListNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SetNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.MapNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	}
	return attributeStatus{}, false
}

func nestedAttributes(attr schema.Attribute) map[string]schema.Attribute {
	switch a := attr.(type) {
	case schema.SingleNestedAttribute:
		return a.Attributes
	case schema.ListNestedAttribute:
		return a.NestedObject.Attributes
	case schema.SetNestedAttribute:
		return a.NestedObject.Attributes
	case schema.MapNestedAttribute:
		return a.NestedObject.Attributes
	}
	return nil
}

// walkAttributes recurses over attrs in alphabetical order so test output is
// deterministic. Paths use dot-separated notation.
func walkAttributes(prefix string, attrs map[string]schema.Attribute, visit func(path string, attr schema.Attribute)) {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		attr := attrs[k]
		visit(path, attr)
		if children := nestedAttributes(attr); children != nil {
			walkAttributes(path, children, visit)
		}
	}
}

type resourceCase struct {
	typeName string
	resource resource.Resource
}

func allResourceCases(t *testing.T) []resourceCase {
	t.Helper()
	var p qProvider
	ctx := context.Background()
	ctors := p.Resources(ctx)

	cases := make([]resourceCase, 0, len(ctors))
	for _, ctor := range ctors {
		r := ctor()
		mdResp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "qovery"}, mdResp)
		if mdResp.TypeName == "" {
			t.Fatalf("resource %T returned an empty TypeName from Metadata; cannot test", r)
		}
		cases = append(cases, resourceCase{typeName: mdResp.TypeName, resource: r})
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].typeName < cases[j].typeName })
	return cases
}

func schemaFor(t *testing.T, r resource.Resource) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestRegression_PlanNoise_NoComputedFlicker(t *testing.T) {
	t.Parallel()

	for _, tc := range allResourceCases(t) {
		tc := tc
		t.Run(tc.typeName, func(t *testing.T) {
			t.Parallel()

			sch := schemaFor(t, tc.resource)

			var problems []string
			walkAttributes("", sch.Attributes, func(path string, attr schema.Attribute) {
				status, ok := inspectAttribute(attr)
				if !ok {
					t.Errorf("unhandled attribute type %T at %s.%s — extend inspectAttribute()", attr, tc.typeName, path)
					return
				}
				if !status.computed {
					return
				}
				if status.preservesState || status.hasDefault {
					return
				}
				if _, allowed := flickerAllowlist[tc.typeName+"."+path]; allowed {
					return
				}
				problems = append(problems, path)
			})

			if len(problems) == 0 {
				return
			}

			sort.Strings(problems)
			t.Errorf(
				"%s has %d Computed attribute(s) with no state-preserving plan modifier and no Default.\n"+
					"Add UseStateForUnknown() (or a state-preserving modifier from qovery/plan_modifiers.go) to each,\n"+
					"or add the path to flickerAllowlist with a written reason if the volatility is legitimate.\n\n"+
					"Failing attributes:\n  - %s",
				tc.typeName, len(problems), strings.Join(problems, "\n  - "),
			)
		})
	}
}
