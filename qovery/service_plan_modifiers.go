package qovery

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The modifiers in this file plan the value q-core stores when a service attribute is omitted,
// for attributes whose default cannot be a static schema Default. Planning the default, rather
// than keeping the state value, makes removing the attribute from the configuration plan the
// reset, and makes a value changed outside Terraform show up in the plan.

const (
	// q-core picks a job's default icon from its schedule (JobDomain.getDefaultIconUri).
	jobCronIconURIDefault      = "app://qovery-console/cron-job"
	jobLifecycleIconURIDefault = "app://qovery-console/lifecycle-job"

	// q-core stores GENERIC when a lifecycle job omits its type, and no type on a cron job.
	jobLifecycleTypeDefault = "GENERIC"
)

// portNameDefaultModifier plans the name q-core gives a port whose name is omitted:
// "p<internal_port>" (PortDto.toDomain).
type portNameDefaultModifier struct{}

func (m portNameDefaultModifier) Description(_ context.Context) string {
	return "When the name is omitted, plans the name the Qovery API gives the port: p<internal_port>."
}

func (m portNameDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m portNameDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var internalPort types.Int64
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName("internal_port"), &internalPort)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if internalPort.IsNull() || internalPort.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}
	resp.PlanValue = types.StringValue(fmt.Sprintf("p%d", internalPort.ValueInt64()))
}

// PortNameDefault returns the plan modifier for ports.name on application and container.
func PortNameDefault() planmodifier.String {
	return portNameDefaultModifier{}
}

// jobIconUriDefaultModifier plans q-core's default job icon, which depends on whether the job is
// a cron job or a lifecycle job.
type jobIconUriDefaultModifier struct{}

func (m jobIconUriDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("When the icon is omitted, plans the Qovery default: %s for a cron job, %s for a lifecycle job.", jobCronIconURIDefault, jobLifecycleIconURIDefault)
}

func (m jobIconUriDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m jobIconUriDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	isCron, known, diags := plannedJobScheduleIsCron(ctx, req.Plan, path.Root("schedule"))
	resp.Diagnostics.Append(diags...)
	switch {
	case diags.HasError():
		return
	case !known:
		resp.PlanValue = types.StringUnknown()
	case isCron:
		resp.PlanValue = types.StringValue(jobCronIconURIDefault)
	default:
		resp.PlanValue = types.StringValue(jobLifecycleIconURIDefault)
	}
}

// JobIconUriDefault returns the plan modifier for qovery_job.icon_uri.
func JobIconUriDefault() planmodifier.String {
	return jobIconUriDefaultModifier{}
}

// jobLifecycleTypeDefaultModifier plans the lifecycle type q-core stores when it is omitted:
// GENERIC for a lifecycle job, none for a cron job (JobScheduleRequest.toDomain).
type jobLifecycleTypeDefaultModifier struct{}

func (m jobLifecycleTypeDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("When the lifecycle type is omitted, plans %s for a lifecycle job and null for a cron job.", jobLifecycleTypeDefault)
}

func (m jobLifecycleTypeDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m jobLifecycleTypeDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	isCron, known, diags := plannedJobScheduleIsCron(ctx, req.Plan, req.Path.ParentPath())
	resp.Diagnostics.Append(diags...)
	switch {
	case diags.HasError():
		return
	case !known:
		resp.PlanValue = types.StringUnknown()
	case isCron:
		resp.PlanValue = types.StringNull()
	default:
		resp.PlanValue = types.StringValue(jobLifecycleTypeDefault)
	}
}

// JobLifecycleTypeDefault returns the plan modifier for qovery_job.schedule.lifecycle_type.
func JobLifecycleTypeDefault() planmodifier.String {
	return jobLifecycleTypeDefaultModifier{}
}

// plannedJobScheduleIsCron reports whether the job schedule planned at schedulePath is a cron
// job. q-core treats any schedule with a cronjob as a cron job and every other one as a
// lifecycle job. known is false while the plan cannot tell yet.
func plannedJobScheduleIsCron(ctx context.Context, plan tfsdk.Plan, schedulePath path.Path) (isCron bool, known bool, diags diag.Diagnostics) {
	var schedule types.Object
	diags = plan.GetAttribute(ctx, schedulePath, &schedule)
	if diags.HasError() || schedule.IsNull() || schedule.IsUnknown() {
		return false, false, diags
	}
	cronjob, ok := schedule.Attributes()["cronjob"]
	if !ok || cronjob.IsUnknown() {
		return false, false, diags
	}
	return !cronjob.IsNull(), true, diags
}

// customDomainsBoolDefaultsModifier plans false for the listed bool attributes of every custom
// domain whose configuration omits them.
//
// It works on the whole set instead of being a schema Default on each attribute: the framework
// looks up a set element's configuration by the planned element value, which holds the Computed
// id and status the configuration lacks, so a nested Default would overwrite configured values
// and the planned elements would no longer correlate with the applied ones. Here the elements
// are matched by domain, which is unique per service.
type customDomainsBoolDefaultsModifier struct {
	attributes []string
}

func (m customDomainsBoolDefaultsModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Plans false for %v when a custom domain omits them.", m.attributes)
}

func (m customDomainsBoolDefaultsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m customDomainsBoolDefaultsModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	configByDomain := make(map[string]map[string]attr.Value, len(req.ConfigValue.Elements()))
	for _, element := range req.ConfigValue.Elements() {
		if domain, attributes, ok := customDomainElement(element); ok {
			configByDomain[domain] = attributes
		}
	}

	elements := make([]attr.Value, 0, len(req.PlanValue.Elements()))
	for _, element := range req.PlanValue.Elements() {
		elements = append(elements, m.withDefaults(ctx, element, configByDomain))
	}

	planValue, diags := types.SetValue(req.PlanValue.ElementType(ctx), elements)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	resp.PlanValue = planValue
}

func (m customDomainsBoolDefaultsModifier) withDefaults(ctx context.Context, element attr.Value, configByDomain map[string]map[string]attr.Value) attr.Value {
	domain, planned, ok := customDomainElement(element)
	if !ok {
		return element
	}
	config, ok := configByDomain[domain]
	if !ok {
		return element
	}

	attributes := maps.Clone(planned)
	for _, name := range m.attributes {
		if configValue, ok := config[name]; ok && configValue.IsNull() {
			attributes[name] = types.BoolValue(false)
		}
	}
	return types.ObjectValueMust(element.(types.Object).AttributeTypes(ctx), attributes)
}

// customDomainElement returns the domain and attributes of a known custom domain element.
func customDomainElement(element attr.Value) (string, map[string]attr.Value, bool) {
	object, ok := element.(types.Object)
	if !ok || object.IsNull() || object.IsUnknown() {
		return "", nil, false
	}
	domain, ok := object.Attributes()["domain"].(types.String)
	if !ok || domain.IsNull() || domain.IsUnknown() {
		return "", nil, false
	}
	return domain.ValueString(), object.Attributes(), true
}

// CustomDomainsBoolDefaults returns the plan modifier for custom_domains that defaults the given
// bool attributes to false, the value the provider has always sent when they are omitted.
func CustomDomainsBoolDefaults(attributes ...string) planmodifier.Set {
	return customDomainsBoolDefaultsModifier{attributes: attributes}
}

// useStateUnlessRepositoryChangesModifier keeps the state value of an omitted git branch while
// the repository URL next to it is unchanged. An omitted branch means the repository's default
// branch, which only the API resolves: the provider cannot plan it, so it keeps the resolved one
// instead of planning unknown and letting every update reset it. When the URL changes, the value
// stays unknown and the update sends no branch, so the API resolves the new repository's default.
type useStateUnlessRepositoryChangesModifier struct{}

func (m useStateUnlessRepositoryChangesModifier) Description(_ context.Context) string {
	return "Keeps the branch recorded in state while the repository URL is unchanged."
}

func (m useStateUnlessRepositoryChangesModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateUnlessRepositoryChangesModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.StateValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	urlPath := req.Path.ParentPath().AtName("url")
	var plannedURL, stateURL types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, urlPath, &plannedURL)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, urlPath, &stateURL)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plannedURL.IsUnknown() || !plannedURL.Equal(stateURL) {
		return
	}
	resp.PlanValue = req.StateValue
}

// UseStateUnlessRepositoryChanges returns the plan modifier for git_repository.branch on
// qovery_application and source.git_repository.branch on qovery_helm.
func UseStateUnlessRepositoryChanges() planmodifier.String {
	return useStateUnlessRepositoryChangesModifier{}
}

// rejectChangeAfterCreateModifier rejects a planned change to a value the Qovery API cannot
// change once the service exists, such as blueprint_id, which it ignores on update, or a job's
// lifecycle type, which it rejects. Applying such a change would fail, or report the old value
// back as an "inconsistent result after apply", so the plan fails instead and names the value
// to declare. It runs after the modifiers that plan the value, so it also catches an omitted
// value whose planned default differs from the recorded one.
type rejectChangeAfterCreateModifier struct {
	// reason says why the Qovery API cannot change the value, as a sentence.
	reason string
}

func (m rejectChangeAfterCreateModifier) Description(_ context.Context) string {
	return "Rejects any change to the value after the service is created."
}

func (m rejectChangeAfterCreateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m rejectChangeAfterCreateModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Creation, or destroy: nothing is recorded yet, or nothing is planned.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() || req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		fmt.Sprintf("Cannot change %s after creation", req.Path),
		fmt.Sprintf("%s "+
			"To keep this service, set %s to the value in the Terraform state (%s) in the configuration. "+
			"To use a different value, recreate the service explicitly, e.g. `terraform destroy -target=<resource address>` followed by `terraform apply`. "+
			"Note that `terraform apply -replace=...` cannot be used here because the plan is rejected before the replacement is applied.",
			m.reason, req.Path, req.StateValue),
	)
}

// RejectChangeAfterCreate returns the plan modifier for values the Qovery API cannot change once
// the service exists. reason says why, as a sentence.
func RejectChangeAfterCreate(reason string) planmodifier.String {
	return rejectChangeAfterCreateModifier{reason: reason}
}

// Reasons passed to RejectChangeAfterCreate.
const (
	blueprintIDChangeReason   = "The Qovery API records blueprint_id only when the service is created and ignores later changes."
	lifecycleTypeChangeReason = "The Qovery API cannot change the lifecycle type of an existing job."
)
