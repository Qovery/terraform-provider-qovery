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
	TfVarFiles              types.List                `tfsdk:"tfvar_files"`
	Variables               types.Set                 `tfsdk:"variable"`
	Backend                 *TerraformBackend         `tfsdk:"backend"`
	Engine                  types.String              `tfsdk:"engine"`
	ProviderVersion         *TerraformProviderVersion `tfsdk:"provider_version"`
	JobResources            *TerraformJobResources    `tfsdk:"job_resources"`
	TimeoutSec              types.Int64               `tfsdk:"timeout_sec"`
	IconURI                 types.String              `tfsdk:"icon_uri"`
	UseClusterCredentials   types.Bool                `tfsdk:"use_cluster_credentials"`
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

type TerraformProviderVersion struct {
	ExplicitVersion        types.String `tfsdk:"explicit_version"`
	ReadFromTerraformBlock types.Bool   `tfsdk:"read_from_terraform_block"`
}

type TerraformJobResources struct {
	CPUMilli   types.Int64 `tfsdk:"cpu_milli"`
	RAMMiB     types.Int64 `tfsdk:"ram_mib"`
	GPU        types.Int64 `tfsdk:"gpu"`
	StorageGiB types.Int64 `tfsdk:"storage_gib"`
}

type TerraformVariable struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
}

// toUpsertServiceRequest converts Terraform model to domain service request
func (t TerraformService) toUpsertServiceRequest(state *TerraformService) (*terraformservice.UpsertServiceRequest, error) {
	return &terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: t.toUpsertRepositoryRequest(),
	}, nil
}

// toUpsertRepositoryRequest converts Terraform model to domain repository request
func (t TerraformService) toUpsertRepositoryRequest() terraformservice.UpsertRepositoryRequest {
	return terraformservice.UpsertRepositoryRequest{
		Name:                  ToString(t.Name),
		Description:           ToString(t.Description),
		AutoDeploy:            ToBool(t.AutoDeploy),
		GitRepository:         t.GitRepository.toDomain(),
		TfVarFiles:            ToStringArray(t.TfVarFiles),
		Variables:             toVariableArray(t.Variables),
		Backend:               t.Backend.toDomain(),
		Engine:                terraformservice.Engine(ToString(t.Engine)),
		ProviderVersion:       t.ProviderVersion.toDomain(),
		JobResources:          t.JobResources.toDomain(),
		TimeoutSec:            ToInt32Pointer(t.TimeoutSec),
		IconURI:               ToString(t.IconURI),
		UseClusterCredentials: ToBool(t.UseClusterCredentials),
		ActionExtraArguments:  toActionExtraArguments(t.ActionExtraArguments),
		AdvancedSettingsJson:  ToString(t.AdvancedSettingsJson),
	}
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

// toDomain converts Terraform provider version to domain
func (p *TerraformProviderVersion) toDomain() terraformservice.ProviderVersion {
	if p == nil {
		return terraformservice.ProviderVersion{}
	}

	return terraformservice.ProviderVersion{
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

// toVariableArray converts Terraform variables set to domain array
func toVariableArray(variablesSet types.Set) []terraformservice.Variable {
	if variablesSet.IsNull() || variablesSet.IsUnknown() {
		return []terraformservice.Variable{}
	}

	var tfVars []TerraformVariable
	variablesSet.ElementsAs(context.Background(), &tfVars, false)

	variables := make([]terraformservice.Variable, 0, len(tfVars))
	for _, v := range tfVars {
		variables = append(variables, terraformservice.Variable{
			Key:    ToString(v.Key),
			Value:  ToString(v.Value),
			Secret: ToBool(v.Secret),
		})
	}

	return variables
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
		Description:             FromString(ts.Description),
		AutoDeploy:              FromBool(ts.AutoDeploy),
		GitRepository:           fromGitRepository(ts.GitRepository),
		TfVarFiles:              FromStringArray(ts.TfVarFiles),
		Variables:               fromVariableArray(ctx, ts.Variables),
		Backend:                 fromBackend(ts.Backend),
		Engine:                  FromString(string(ts.Engine)),
		ProviderVersion:         fromProviderVersion(ts.ProviderVersion),
		JobResources:            fromJobResources(ts.JobResources),
		TimeoutSec:              FromInt32Pointer(ts.TimeoutSec),
		IconURI:                 FromString(ts.IconURI),
		UseClusterCredentials:   FromBool(ts.UseClusterCredentials),
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

// fromProviderVersion converts domain provider version to Terraform
func fromProviderVersion(p terraformservice.ProviderVersion) *TerraformProviderVersion {
	return &TerraformProviderVersion{
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

// fromVariableArray converts domain variables to Terraform set
func fromVariableArray(ctx context.Context, variables []terraformservice.Variable) types.Set {
	if len(variables) == 0 {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":    types.StringType,
				"value":  types.StringType,
				"secret": types.BoolType,
			},
		})
	}

	tfVars := make([]attr.Value, 0, len(variables))
	for _, v := range variables {
		objValue, diag := types.ObjectValue(
			map[string]attr.Type{
				"key":    types.StringType,
				"value":  types.StringType,
				"secret": types.BoolType,
			},
			map[string]attr.Value{
				"key":    FromString(v.Key),
				"value":  FromString(v.Value),
				"secret": FromBool(v.Secret),
			},
		)
		if diag.HasError() {
			return types.SetNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"key":    types.StringType,
					"value":  types.StringType,
					"secret": types.BoolType,
				},
			})
		}
		tfVars = append(tfVars, objValue)
	}

	setValue, diag := types.SetValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":    types.StringType,
				"value":  types.StringType,
				"secret": types.BoolType,
			},
		},
		tfVars,
	)
	if diag.HasError() {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":    types.StringType,
				"value":  types.StringType,
				"secret": types.BoolType,
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
