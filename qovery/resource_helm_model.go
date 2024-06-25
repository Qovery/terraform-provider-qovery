package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Helm struct {
	ID                           types.String         `tfsdk:"id"`
	EnvironmentID                types.String         `tfsdk:"environment_id"`
	Name                         types.String         `tfsdk:"name"`
	TimeoutSec                   types.Int64          `tfsdk:"timeout_sec"`
	AutoPreview                  types.Bool           `tfsdk:"auto_preview"`
	AutoDeploy                   types.Bool           `tfsdk:"auto_deploy"`
	Arguments                    types.Set            `tfsdk:"arguments"`
	AllowClusterWideResources    types.Bool           `tfsdk:"allow_cluster_wide_resources"`
	Source                       *HelmSource          `tfsdk:"source"`
	ValuesOverride               *HelmValuesOverride  `tfsdk:"values_override"`
	Ports                        *map[string]HelmPort `tfsdk:"ports"`
	BuiltInEnvironmentVariables  types.Set            `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set            `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set            `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set            `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set            `tfsdk:"secrets"`
	SecretAliases                types.Set            `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set            `tfsdk:"secret_overrides"`
	ExternalHost                 types.String         `tfsdk:"external_host"`
	InternalHost                 types.String         `tfsdk:"internal_host"`
	DeploymentStageId            types.String         `tfsdk:"deployment_stage_id"`
	AdvancedSettingsJson         types.String         `tfsdk:"advanced_settings_json"`
	DeploymentRestrictions       types.Set            `tfsdk:"deployment_restrictions"`
}

type HelmSource struct {
	HelmSourceHelmRepository *HelmSourceHelmRepository `tfsdk:"helm_repository"`
	HelmSourceGitRepository  *HelmSourceGitRepository  `tfsdk:"git_repository"`
}

type HelmSourceHelmRepository struct {
	HelmRepositoryId types.String `tfsdk:"helm_repository_id"`
	HelmChartName    types.String `tfsdk:"chart_name"`
	HelmChartVersion types.String `tfsdk:"chart_version"`
}

type HelmSourceGitRepository struct {
	Url        types.String `tfsdk:"url"`
	Branch     types.String `tfsdk:"branch"`
	RootPath   types.String `tfsdk:"root_path"`
	GitTokenId types.String `tfsdk:"git_token_id"`
}

type HelmValuesOverride struct {
	HelmValuesOverrideSet       types.Map               `tfsdk:"set"`
	HelmValuesOverrideSetString types.Map               `tfsdk:"set_string"`
	HelmValuesOverrideSetJson   types.Map               `tfsdk:"set_json"`
	HelmValuesOverrideFile      *HelmValuesOverrideFile `tfsdk:"file"`
}

type HelmValuesOverrideFile struct {
	Raw           *map[string]HelmValuesRaw `tfsdk:"raw"`
	GitRepository *HelmValuesGitRepository  `tfsdk:"git_repository"`
}

type HelmValuesRaw struct {
	Content types.String `tfsdk:"content"`
}

type HelmValuesGitRepository struct {
	Url        types.String `tfsdk:"url"`
	Branch     types.String `tfsdk:"branch"`
	Paths      types.Set    `tfsdk:"paths"`
	GitTokenId types.String `tfsdk:"git_token_id"`
}

type HelmPort struct {
	ServiceName  types.String `tfsdk:"service_name"`
	Namespace    types.String `tfsdk:"namespace"`
	InternalPort types.Int64  `tfsdk:"internal_port"`
	ExternalPort types.Int64  `tfsdk:"external_port"`
	Protocol     types.String `tfsdk:"protocol"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
}

func (h Helm) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(h.EnvironmentVariables)
}
func (h Helm) EnvironmentVariableAliasesList() EnvironmentVariableList {
	return toEnvironmentVariableList(h.EnvironmentVariableAliases)
}
func (h Helm) EnvironmentVariableOverridesList() EnvironmentVariableList {
	return toEnvironmentVariableList(h.EnvironmentVariableOverrides)
}

func (h Helm) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(h.BuiltInEnvironmentVariables)
}

func (h Helm) SecretList() SecretList {
	return ToSecretList(h.Secrets)
}
func (h Helm) SecretAliasesList() SecretList {
	return ToSecretList(h.SecretAliases)
}
func (h Helm) SecretOverridesList() SecretList {
	return ToSecretList(h.SecretOverrides)
}

func (j Helm) DeploymentRestrictionDiff(deploymentRestrictionsState *types.Set) (*deploymentrestriction.ServiceDeploymentRestrictionsDiff, error) {
	return deploymentrestriction.ToDeploymentRestrictionDiff(j.DeploymentRestrictions, deploymentRestrictionsState)
}

func (h Helm) toUpsertServiceRequest(state *Helm) (*helm.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	var stateEnvironmentVariableAliases EnvironmentVariableList
	var stateEnvironmentVariableOverrides EnvironmentVariableList
	var stateSecrets SecretList
	var stateSecretAliases SecretList
	var stateSecretOverrides SecretList
	var stateDeploymentRestrictions types.Set

	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
		stateEnvironmentVariableAliases = state.EnvironmentVariableAliasesList()
		stateEnvironmentVariableOverrides = state.EnvironmentVariableOverridesList()
		stateSecrets = state.SecretList()
		stateSecretAliases = state.SecretAliasesList()
		stateSecretOverrides = state.SecretOverridesList()
		stateDeploymentRestrictions = state.DeploymentRestrictions
	}

	helmRequest, err := h.toUpsertRepositoryRequest()
	if err != nil {
		return nil, err
	}

	deploymentRestrictionsDiff, err := h.DeploymentRestrictionDiff(&stateDeploymentRestrictions)
	if err != nil {
		return nil, err
	}

	return &helm.UpsertServiceRequest{
		HelmUpsertRequest:            *helmRequest,
		EnvironmentVariables:         h.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		EnvironmentVariableAliases:   h.EnvironmentVariableAliasesList().diffRequest(stateEnvironmentVariableAliases),
		EnvironmentVariableOverrides: h.EnvironmentVariableOverridesList().diffRequest(stateEnvironmentVariableOverrides),
		Secrets:                      h.SecretList().diffRequest(stateSecrets),
		SecretAliases:                h.SecretAliasesList().diffRequest(stateSecretAliases),
		SecretOverrides:              h.SecretOverridesList().diffRequest(stateSecretOverrides),
		DeploymentRestrictionsDiff:   *deploymentRestrictionsDiff,
	}, nil
}

func (h Helm) toUpsertRepositoryRequest() (*helm.UpsertRepositoryRequest, error) {
	var ports *[]helm.Port = nil
	if h.Ports != nil {
		portLists := make([]helm.Port, 0, len(*h.Ports))
		for portName, value := range *h.Ports {
			protocol, err := helm.NewProtocolFromString(ToString(value.Protocol))
			if err != nil {
				return nil, errors.Wrap(err, port.ErrInvalidProtocolParam.Error())
			}

			helmPort := helm.Port{
				Name:         portName,
				InternalPort: ToInt32(value.InternalPort),
				ExternalPort: ToInt32Pointer(value.ExternalPort),
				ServiceName:  ToString(value.ServiceName),
				Namespace:    ToStringPointer(value.Namespace),
				Protocol:     *protocol,
				IsDefault:    ToBool(value.IsDefault),
			}

			portLists = append(portLists, helmPort)
		}
		ports = &portLists
	}

	return &helm.UpsertRepositoryRequest{
		Name:                      ToString(h.Name),
		TimeoutSec:                ToInt32Pointer(h.TimeoutSec),
		AutoPreview:               *qovery.NewNullableBool(ToBoolPointer(h.AutoPreview)),
		AutoDeploy:                ToBool(h.AutoDeploy),
		Arguments:                 ToStringArrayFromSet(h.Arguments),
		AllowClusterWideResources: ToBool(h.AllowClusterWideResources),
		Source:                    h.Source.toUpsertRequest(),
		ValuesOverride:            h.ValuesOverride.toUpsertRequest(),
		DeploymentStageID:         ToString(h.DeploymentStageId),
		Ports:                     ports,
		AdvancedSettingsJson:      ToString(h.AdvancedSettingsJson),
	}, nil
}

func (source HelmSource) toUpsertRequest() helm.Source {
	return helm.Source{
		GitRepository:  source.HelmSourceGitRepository.toUpsertRequest(),
		HelmRepository: source.HelmSourceHelmRepository.toUpsertRequest(),
	}
}

func (r *HelmSourceGitRepository) toUpsertRequest() *helm.SourceGitRepository {
	if r == nil {
		return nil
	}

	return &helm.SourceGitRepository{
		Url:        ToString(r.Url),
		Branch:     ToStringPointer(r.Branch),
		RootPath:   ToString(r.RootPath),
		GitTokenId: ToStringPointer(r.GitTokenId),
	}
}

func (r *HelmSourceHelmRepository) toUpsertRequest() *helm.SourceHelmRepository {
	if r == nil {
		return nil
	}

	return &helm.SourceHelmRepository{
		RepositoryId: ToString(r.HelmRepositoryId),
		ChartName:    ToString(r.HelmChartName),
		ChartVersion: ToString(r.HelmChartVersion),
	}
}

func (f HelmValuesOverrideFile) toUpsertRequest() helm.ValuesOverrideFile {
	var raw *helm.Raw = nil
	if f.Raw != nil {
		values := make([]helm.RawValue, 0, len(*f.Raw))
		for key, value := range *f.Raw {
			rawValue := helm.RawValue{
				Name:    key,
				Content: value.Content.ValueString(),
			}

			values = append(values, rawValue)
		}

		raw = &helm.Raw{Values: values}
	}

	var gitRepository *helm.ValuesOverrideGit
	if f.GitRepository != nil {
		paths := make([]string, 0, len(f.GitRepository.Paths.Elements()))
		for _, elem := range f.GitRepository.Paths.Elements() {
			paths = append(paths, elem.(types.String).ValueString())
		}

		var gitTokenId *string = nil
		if !f.GitRepository.GitTokenId.IsNull() {
			v := ToStringPointer(f.GitRepository.GitTokenId)
			gitTokenId = v
		}

		gitRepository = &helm.ValuesOverrideGit{
			Url:      f.GitRepository.Url.ValueString(),
			Branch:   f.GitRepository.Branch.ValueString(),
			Paths:    paths,
			GitToken: gitTokenId,
		}
	}

	return helm.ValuesOverrideFile{
		Raw:           raw,
		GitRepository: gitRepository,
	}
}

func convertSetElements(elements map[string]attr.Value) [][]string {
	set := make([][]string, 0, len(elements))

	for key, value := range elements {
		set = append(set, []string{key, value.(types.String).ValueString()})
	}

	return set
}

func (valuesOverride HelmValuesOverride) toUpsertRequest() helm.ValuesOverride {
	set := convertSetElements(valuesOverride.HelmValuesOverrideSet.Elements())
	setString := convertSetElements(valuesOverride.HelmValuesOverrideSetString.Elements())
	setJson := convertSetElements(valuesOverride.HelmValuesOverrideSetJson.Elements())

	var file *helm.ValuesOverrideFile
	if valuesOverride.HelmValuesOverrideFile != nil {
		valuesOverrideFile := (*valuesOverride.HelmValuesOverrideFile).toUpsertRequest()
		file = &valuesOverrideFile
	}

	return helm.ValuesOverride{
		Set:       set,
		SetString: setString,
		SetJson:   setJson,
		File:      file,
	}
}

func convertSetToHelmValuesOverrideSet(ctx context.Context, set [][]string, state *types.Map) types.Map {
	if state != nil && len(set) == 0 && (state.IsUnknown() || state.IsNull()) {
		return *state
	}

	elements := make(map[string]string, len(set))

	for _, kv := range set {
		if len(kv) == 2 {
			elements[kv[0]] = kv[1]
		} else {
			fmt.Println("Invalid key-value pair:", kv)
		}
	}

	helmValuesOverrideSet, diagnostics := types.MapValueFrom(ctx, types.StringType, elements)
	if diagnostics.HasError() {
		panic("Error during helm values override conversion")
	}

	return helmValuesOverrideSet
}

func HelmValuesOverrideFromDomainHelmValuesOverride(ctx context.Context, h helm.ValuesOverride, state *HelmValuesOverride) HelmValuesOverride {
	var helmValuesOverrideSet types.Map
	var helmValuesOverrideSetString types.Map
	var helmValuesOverrideSetJson types.Map
	if state == nil {
		helmValuesOverrideSet = convertSetToHelmValuesOverrideSet(ctx, h.Set, nil)
		helmValuesOverrideSetString = convertSetToHelmValuesOverrideSet(ctx, h.SetString, nil)
		helmValuesOverrideSetJson = convertSetToHelmValuesOverrideSet(ctx, h.SetJson, nil)
	} else {
		helmValuesOverrideSet = convertSetToHelmValuesOverrideSet(ctx, h.Set, &state.HelmValuesOverrideSetString)
		helmValuesOverrideSetString = convertSetToHelmValuesOverrideSet(ctx, h.SetString, &state.HelmValuesOverrideSetString)
		helmValuesOverrideSetJson = convertSetToHelmValuesOverrideSet(ctx, h.SetJson, &state.HelmValuesOverrideSetJson)
	}

	var gitRepository *HelmValuesGitRepository
	if h.File.GitRepository != nil {
		gitToken := ""
		if h.File.GitRepository.GitToken != nil {
			gitToken = *h.File.GitRepository.GitToken
		}

		gitRepository = &HelmValuesGitRepository{
			Url:        FromString(h.File.GitRepository.Url),
			Branch:     FromString(h.File.GitRepository.Branch),
			Paths:      FromStringSet(h.File.GitRepository.Paths),
			GitTokenId: FromString(gitToken),
		}
	}

	var raw *map[string]HelmValuesRaw
	if h.File.Raw != nil {
		if (state != nil && state.HelmValuesOverrideFile != nil) || len(h.File.Raw.Values) != 0 {
			helmValuesRaw := make(map[string]HelmValuesRaw, len(h.File.Raw.Values))

			for _, value := range h.File.Raw.Values {
				helmValuesRaw[value.Name] = HelmValuesRaw{Content: FromString(value.Content)}
			}

			raw = &helmValuesRaw
		}
	}

	var helmValuesOverrideFile *HelmValuesOverrideFile
	if raw != nil || gitRepository != nil {
		helmValuesOverrideFile = &HelmValuesOverrideFile{
			GitRepository: gitRepository,
			Raw:           raw,
		}
	}

	return HelmValuesOverride{
		HelmValuesOverrideSet:       helmValuesOverrideSet,
		HelmValuesOverrideSetString: helmValuesOverrideSetString,
		HelmValuesOverrideSetJson:   helmValuesOverrideSetJson,
		HelmValuesOverrideFile:      helmValuesOverrideFile,
	}
}

func HelmSourceFromDomainHelmSource(source helm.Source) HelmSource {
	var helmSourceGitRepository *HelmSourceGitRepository = nil
	if source.GitRepository != nil {

		helmSourceGitRepository = &HelmSourceGitRepository{
			Url:        FromString(source.GitRepository.Url),
			Branch:     FromStringPointer(source.GitRepository.Branch),
			RootPath:   FromString(source.GitRepository.RootPath),
			GitTokenId: FromStringPointer(source.GitRepository.GitTokenId),
		}
	}

	var helmSourceHelmRepository *HelmSourceHelmRepository
	if source.HelmRepository != nil {

		helmSourceHelmRepository = &HelmSourceHelmRepository{
			HelmRepositoryId: FromString(source.HelmRepository.RepositoryId),
			HelmChartName:    FromString(source.HelmRepository.ChartName),
			HelmChartVersion: FromString(source.HelmRepository.ChartVersion),
		}
	}

	return HelmSource{
		HelmSourceGitRepository:  helmSourceGitRepository,
		HelmSourceHelmRepository: helmSourceHelmRepository,
	}
}

func HelmPortsFromDomainHelmPorts(ports []helm.Port) *map[string]HelmPort {
	if len(ports) == 0 {
		return nil
	}

	portsAsMap := make(map[string]HelmPort, len(ports))
	for _, p := range ports {
		portsAsMap[p.Name] = HelmPort{
			ServiceName:  FromString(p.ServiceName),
			Namespace:    FromStringPointer(p.Namespace),
			InternalPort: FromInt32(p.InternalPort),
			ExternalPort: FromInt32Pointer(p.ExternalPort),
			Protocol:     FromString(p.Protocol.String()),
			IsDefault:    FromBool(p.IsDefault),
		}
	}

	return &portsAsMap
}

func convertDomainHelmToHelm(ctx context.Context, state Helm, helm *helm.Helm) Helm {
	source := HelmSourceFromDomainHelmSource(helm.Source)
	valuesOverride := HelmValuesOverrideFromDomainHelmValuesOverride(ctx, helm.ValuesOverride, state.ValuesOverride)
	ports := HelmPortsFromDomainHelmPorts(helm.Ports)

	return Helm{
		ID:                           FromString(helm.ID.String()),
		EnvironmentID:                FromString(helm.EnvironmentID.String()),
		Name:                         FromString(helm.Name),
		TimeoutSec:                   FromInt32Pointer(helm.TimeoutSec),
		AutoPreview:                  FromBool(helm.AutoPreview),
		AutoDeploy:                   FromBool(helm.AutoDeploy),
		Arguments:                    FromStringSet(helm.Arguments),
		AllowClusterWideResources:    FromBool(helm.AllowClusterWideResources),
		Source:                       &source,
		ValuesOverride:               &valuesOverride,
		Ports:                        ports,
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(helm.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformSet(ctx),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariables, helm.EnvironmentVariables, variable.ScopeHelm, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableAliases, helm.EnvironmentVariables, variable.ScopeHelm, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableOverrides, helm.EnvironmentVariables, variable.ScopeHelm, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, helm.Secrets, variable.ScopeHelm, "VALUE").toTerraformSet(ctx),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, helm.Secrets, variable.ScopeHelm, "ALIAS").toTerraformSet(ctx),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, helm.Secrets, variable.ScopeHelm, "OVERRIDE").toTerraformSet(ctx),
		InternalHost:                 FromStringPointer(helm.InternalHost),
		ExternalHost:                 FromStringPointer(helm.ExternalHost),
		DeploymentStageId:            FromString(helm.DeploymentStageID),
		AdvancedSettingsJson:         FromString(helm.AdvancedSettingsJson),
		DeploymentRestrictions:       FromDeploymentRestrictionList(state.DeploymentRestrictions, helm.JobDeploymentRestrictions),
	}
}
