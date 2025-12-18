package qovery

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

type TerraformService struct {
	ID                      types.String              `tfsdk:"id"`
	EnvironmentID           types.String              `tfsdk:"environment_id"`
	Name                    types.String              `tfsdk:"name"`
	Description             types.String              `tfsdk:"description"`
	AutoDeploy              types.Bool                `tfsdk:"auto_deploy"`
	GitRepository           *TerraformGitRepository   `tfsdk:"git_repository"`
	TfVarFiles              types.List                `tfsdk:"tfvars_files"`
	Variables               types.Set                 `tfsdk:"variables"`
	Backend                 *TerraformBackend       `tfsdk:"backend"`
	Engine                  types.String            `tfsdk:"engine"`
	EngineVersion           *TerraformEngineVersion `tfsdk:"engine_version"`
	JobResources            *TerraformJobResources  `tfsdk:"job_resources"`
	TimeoutSec              types.Int64               `tfsdk:"timeout_seconds"`
	IconURI                 types.String              `tfsdk:"icon_uri"`
	UseClusterCredentials   types.Bool                `tfsdk:"use_cluster_credentials"`
	DockerfileFragment      *TerraformDockerfileFragment  `tfsdk:"dockerfile_fragment"`
	ActionExtraArguments    types.Map                 `tfsdk:"action_extra_arguments"`
	AdvancedSettingsJson    types.String              `tfsdk:"advanced_settings_json"`
	CreatedAt               types.String              `tfsdk:"created_at"`
	UpdatedAt               types.String              `tfsdk:"updated_at"`
}

type TerraformGitRepository struct {
	URL        types.String `tfsdk:"url"`
	Branch     types.String `tfsdk:"branch"`
	RootPath   types.String `tfsdk:"root_path"`
	GitTokenID types.String `tfsdk:"git_token_id"`
}

type TerraformBackend struct {
	Kubernetes   *TerraformKubernetesBackend   `tfsdk:"kubernetes"`
	UserProvided *TerraformUserProvidedBackend `tfsdk:"user_provided"`
}

type TerraformKubernetesBackend struct{}

type TerraformUserProvidedBackend struct{}

type TerraformEngineVersion struct {
	ExplicitVersion        types.String `tfsdk:"explicit_version"`
	ReadFromTerraformBlock types.Bool   `tfsdk:"read_from_terraform_block"`
}

type TerraformJobResources struct {
	CPUMilli   types.Int64 `tfsdk:"cpu_milli"`
	RAMMiB     types.Int64 `tfsdk:"ram_mib"`
	GPU        types.Int64 `tfsdk:"gpu"`
	StorageGiB types.Int64 `tfsdk:"storage_gib"`
}

type TerraformDockerfileFragment struct {
	File   *TerraformDockerfileFragmentFile   `tfsdk:"file"`
	Inline *TerraformDockerfileFragmentInline `tfsdk:"inline"`
}

type TerraformDockerfileFragmentFile struct {
	Path types.String `tfsdk:"path"`
}

type TerraformDockerfileFragmentInline struct {
	Content types.String `tfsdk:"content"`
}

type TerraformVariable struct {
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
}

// toUpsertServiceRequest converts Terraform model to domain service request
func (t TerraformService) toUpsertServiceRequest(state *TerraformService) (*terraformservice.UpsertServiceRequest, error) {
	req, err := t.toUpsertRepositoryRequest()
	if err != nil {
		return nil, err
	}

	return &terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: req,
	}, nil
}

// toUpsertRepositoryRequest converts Terraform model to domain repository request
func (t TerraformService) toUpsertRepositoryRequest() (terraformservice.UpsertRepositoryRequest, error) {
	variables, err := toVariableArray(t.Variables)
	if err != nil {
		return terraformservice.UpsertRepositoryRequest{}, err
	}

	return terraformservice.UpsertRepositoryRequest{
		Name:                  ToString(t.Name),
		Description:           ToStringPointer(t.Description),
		AutoDeploy:            ToBool(t.AutoDeploy),
		GitRepository:         t.GitRepository.toDomain(),
		TfVarFiles:            ToStringArray(t.TfVarFiles),
		Variables:             variables,
		Backend:               t.Backend.toDomain(),
		Engine:                terraformservice.Engine(ToString(t.Engine)),
		EngineVersion:         t.EngineVersion.toDomain(),
		JobResources:          t.JobResources.toDomain(),
		TimeoutSec:            ToInt32Pointer(t.TimeoutSec),
		IconURI:               ToString(t.IconURI),
		UseClusterCredentials: ToBool(t.UseClusterCredentials),
		DockerfileFragment:    t.DockerfileFragment.toDomain(),
		ActionExtraArguments:  toActionExtraArguments(t.ActionExtraArguments),
		AdvancedSettingsJson:  ToString(t.AdvancedSettingsJson),
	}, nil
}

// toDomain converts Terraform git repository to domain
func (g *TerraformGitRepository) toDomain() terraformservice.GitRepository {
	if g == nil {
		return terraformservice.GitRepository{}
	}

	var gitTokenID *uuid.UUID
	if !g.GitTokenID.IsNull() && !g.GitTokenID.IsUnknown() {
		tokenID, err := uuid.Parse(ToString(g.GitTokenID))
		if err == nil {
			gitTokenID = &tokenID
		}
	}

	return terraformservice.GitRepository{
		URL:        ToString(g.URL),
		Branch:     ToString(g.Branch),
		RootPath:   ToString(g.RootPath),
		GitTokenID: gitTokenID,
	}
}

// toDomain converts Terraform backend to domain
func (b *TerraformBackend) toDomain() terraformservice.Backend {
	if b == nil {
		return terraformservice.Backend{}
	}

	backend := terraformservice.Backend{}
	if b.Kubernetes != nil {
		backend.Kubernetes = &terraformservice.KubernetesBackend{}
	}
	if b.UserProvided != nil {
		backend.UserProvided = &terraformservice.UserProvidedBackend{}
	}

	return backend
}

// toDomain converts Terraform engine version to domain
func (p *TerraformEngineVersion) toDomain() terraformservice.EngineVersion {
	if p == nil {
		return terraformservice.EngineVersion{}
	}

	return terraformservice.EngineVersion{
		ExplicitVersion:        ToString(p.ExplicitVersion),
		ReadFromTerraformBlock: ToBool(p.ReadFromTerraformBlock),
	}
}

// toDomain converts Terraform job resources to domain
func (j *TerraformJobResources) toDomain() terraformservice.JobResources {
	if j == nil {
		return terraformservice.JobResources{
			CPUMilli:   terraformservice.DefaultCPU,
			RAMMiB:     terraformservice.DefaultRAM,
			GPU:        terraformservice.DefaultGPU,
			StorageGiB: terraformservice.DefaultStorage,
		}
	}

	return terraformservice.JobResources{
		CPUMilli:   ToInt32(j.CPUMilli),
		RAMMiB:     ToInt32(j.RAMMiB),
		GPU:        ToInt32(j.GPU),
		StorageGiB: ToInt32(j.StorageGiB),
	}
}

// toDomain converts Terraform dockerfile fragment to domain
func (d *TerraformDockerfileFragment) toDomain() *terraformservice.DockerfileFragment {
	if d == nil {
		return nil
	}

	fragment := &terraformservice.DockerfileFragment{}
	if d.File != nil {
		fragment.File = &terraformservice.DockerfileFragmentFile{
			Path: ToString(d.File.Path),
		}
	}
	if d.Inline != nil {
		fragment.Inline = &terraformservice.DockerfileFragmentInline{
			Content: ToString(d.Inline.Content),
		}
	}

	// Return nil if neither is set
	if fragment.File == nil && fragment.Inline == nil {
		return nil
	}

	return fragment
}

// toVariableArray converts Terraform variables set to domain array
func toVariableArray(variablesSet types.Set) ([]terraformservice.Variable, error) {
	if variablesSet.IsNull() || variablesSet.IsUnknown() {
		return []terraformservice.Variable{}, nil
	}

	var tfVars []TerraformVariable
	if diags := variablesSet.ElementsAs(context.Background(), &tfVars, false); diags.HasError() {
		return nil, errors.Errorf("failed to convert terraform variables: %v", diags)
	}

	variables := make([]terraformservice.Variable, 0, len(tfVars))
	for _, v := range tfVars {
		variables = append(variables, terraformservice.Variable{
			Key:    ToString(v.Key),
			Value:  ToString(v.Value),
			Secret: ToBool(v.IsSecret),
		})
	}

	return variables, nil
}

// toActionExtraArguments converts Terraform map to domain map
func toActionExtraArguments(argsMap types.Map) map[string][]string {
	if argsMap.IsNull() || argsMap.IsUnknown() {
		return map[string][]string{}
	}

	result := make(map[string][]string)
	elements := argsMap.Elements()

	for key, value := range elements {
		if listValue, ok := value.(types.List); ok && !listValue.IsNull() && !listValue.IsUnknown() {
			args := ToStringArray(listValue)
			result[key] = args
		}
	}

	return result
}

// convertDomainTerraformServiceToTerraformService converts domain entity to Terraform model
func convertDomainTerraformServiceToTerraformService(ctx context.Context, plan TerraformService, ts *terraformservice.TerraformService) TerraformService {
	return TerraformService{
		ID:                      FromString(ts.ID.String()),
		EnvironmentID:           FromString(ts.EnvironmentID.String()),
		Name:                    FromString(ts.Name),
		Description:             FromStringPointer(ts.Description),
		AutoDeploy:              FromBool(ts.AutoDeploy),
		GitRepository:           fromGitRepository(ts.GitRepository),
		TfVarFiles:              FromStringArray(ts.TfVarFiles),
		Variables:               fromVariableArray(ctx, plan.Variables, ts.Variables),
		Backend:                 fromBackend(ts.Backend),
		Engine:                  FromString(string(ts.Engine)),
		EngineVersion:           fromEngineVersion(ts.EngineVersion),
		JobResources:            fromJobResources(ts.JobResources),
		TimeoutSec:              FromInt32Pointer(ts.TimeoutSec),
		IconURI:                 FromString(ts.IconURI),
		UseClusterCredentials:   FromBool(ts.UseClusterCredentials),
		DockerfileFragment:      fromDockerfileFragment(ts.DockerfileFragment),
		ActionExtraArguments:    fromActionExtraArguments(ctx, ts.ActionExtraArguments),
		AdvancedSettingsJson:    FromString(ts.AdvancedSettingsJson),
		CreatedAt:               FromTime(ts.CreatedAt),
		UpdatedAt:               FromTimePointer(ts.UpdatedAt),
	}
}

// fromGitRepository converts domain git repository to Terraform
func fromGitRepository(g terraformservice.GitRepository) *TerraformGitRepository {
	var gitTokenID types.String
	if g.GitTokenID != nil {
		gitTokenID = FromString(g.GitTokenID.String())
	} else {
		gitTokenID = types.StringNull()
	}

	return &TerraformGitRepository{
		URL:        FromString(g.URL),
		Branch:     FromString(g.Branch),
		RootPath:   FromString(g.RootPath),
		GitTokenID: gitTokenID,
	}
}

// fromBackend converts domain backend to Terraform
func fromBackend(b terraformservice.Backend) *TerraformBackend {
	backend := &TerraformBackend{}

	if b.Kubernetes != nil {
		backend.Kubernetes = &TerraformKubernetesBackend{}
	}
	if b.UserProvided != nil {
		backend.UserProvided = &TerraformUserProvidedBackend{}
	}

	return backend
}

// fromEngineVersion converts domain engine version to Terraform
func fromEngineVersion(p terraformservice.EngineVersion) *TerraformEngineVersion {
	return &TerraformEngineVersion{
		ExplicitVersion:        FromString(p.ExplicitVersion),
		ReadFromTerraformBlock: FromBool(p.ReadFromTerraformBlock),
	}
}

// fromJobResources converts domain job resources to Terraform
func fromJobResources(j terraformservice.JobResources) *TerraformJobResources {
	return &TerraformJobResources{
		CPUMilli:   FromInt32(j.CPUMilli),
		RAMMiB:     FromInt32(j.RAMMiB),
		GPU:        FromInt32(j.GPU),
		StorageGiB: FromInt32(j.StorageGiB),
	}
}

// fromDockerfileFragment converts domain dockerfile fragment to Terraform
func fromDockerfileFragment(d *terraformservice.DockerfileFragment) *TerraformDockerfileFragment {
	if d == nil {
		return nil
	}

	fragment := &TerraformDockerfileFragment{}
	if d.File != nil {
		fragment.File = &TerraformDockerfileFragmentFile{
			Path: FromString(d.File.Path),
		}
	}
	if d.Inline != nil {
		fragment.Inline = &TerraformDockerfileFragmentInline{
			Content: FromString(d.Inline.Content),
		}
	}

	// Return nil if neither is set
	if fragment.File == nil && fragment.Inline == nil {
		return nil
	}

	return fragment
}

// fromVariableArray converts domain variables to Terraform set while preserving sensitive values from plan
func fromVariableArray(ctx context.Context, planVars types.Set, variables []terraformservice.Variable) types.Set {
	// If API returns no variables but plan has variables, preserve the plan (API might not return them)
	if len(variables) == 0 {
		if !planVars.IsNull() && !planVars.IsUnknown() && len(planVars.Elements()) > 0 {
			return planVars
		}
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":       types.StringType,
				"value":     types.StringType,
				"is_secret": types.BoolType,
			},
		})
	}

	// Build a map of existing variables from plan state (keyed by variable key)
	planVarMap := make(map[string]types.Object)
	if !planVars.IsNull() && !planVars.IsUnknown() {
		planVarElements := planVars.Elements()
		for _, elem := range planVarElements {
			if objVal, ok := elem.(types.Object); ok {
				attrs := objVal.Attributes()
				if keyAttr, exists := attrs["key"]; exists {
					if keyStr, ok := keyAttr.(types.String); ok && !keyStr.IsNull() {
						planVarMap[keyStr.ValueString()] = objVal
					}
				}
			}
		}
	}

	// Process all variables from plan first (to ensure we include any that were in plan but not in API response)
	processedKeys := make(map[string]bool)
	tfVars := make([]attr.Value, 0, len(variables))

	// For each variable from the API, check if it exists in plan and use plan's value
	for _, v := range variables {
		varValue := FromString(v.Value)

		// Always check plan for this variable to preserve sensitive values
		if planVar, exists := planVarMap[v.Key]; exists {
			attrs := planVar.Attributes()
			if valueAttr, ok := attrs["value"]; ok {
				if valueStr, ok := valueAttr.(types.String); ok {
					// Preserve the value from plan to maintain sensitivity
					varValue = valueStr
				}
			}
		}

		objValue, diag := types.ObjectValue(
			map[string]attr.Type{
				"key":       types.StringType,
				"value":     types.StringType,
				"is_secret": types.BoolType,
			},
			map[string]attr.Value{
				"key":       FromString(v.Key),
				"value":     varValue,
				"is_secret": FromBool(v.Secret),
			},
		)
		if diag.HasError() {
			return types.SetNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"key":       types.StringType,
					"value":     types.StringType,
					"is_secret": types.BoolType,
				},
			})
		}
		tfVars = append(tfVars, objValue)
		processedKeys[v.Key] = true
	}

	setValue, diag := types.SetValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":       types.StringType,
				"value":     types.StringType,
				"is_secret": types.BoolType,
			},
		},
		tfVars,
	)
	if diag.HasError() {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":       types.StringType,
				"value":     types.StringType,
				"is_secret": types.BoolType,
			},
		})
	}

	return setValue
}

// fromActionExtraArguments converts domain map to Terraform map
func fromActionExtraArguments(ctx context.Context, args map[string][]string) types.Map {
	if len(args) == 0 {
		return types.MapNull(types.ListType{ElemType: types.StringType})
	}

	elements := make(map[string]attr.Value)
	for key, values := range args {
		listValue := FromStringArray(values)
		elements[key] = listValue
	}

	mapValue, diag := types.MapValue(types.ListType{ElemType: types.StringType}, elements)
	if diag.HasError() {
		return types.MapNull(types.ListType{ElemType: types.StringType})
	}

	return mapValue
}

// Validate validates the terraform service model
func (t TerraformService) Validate() error {
	// Validate that exactly one backend type is set
	if t.Backend == nil {
		return errors.New("backend is required")
	}

	hasKubernetes := t.Backend.Kubernetes != nil
	hasUserProvided := t.Backend.UserProvided != nil

	if !hasKubernetes && !hasUserProvided {
		return errors.New("exactly one backend type must be specified: kubernetes or user_provided")
	}

	if hasKubernetes && hasUserProvided {
		return errors.New("cannot specify both kubernetes and user_provided backend types")
	}

	return nil
}
