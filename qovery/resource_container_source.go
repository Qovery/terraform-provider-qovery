package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

const containerGitSourceRootPathDefault = "/"

var _ resource.ConfigValidator = containerSourceValidator{}

type ContainerSource struct {
	GitRepository *ContainerSourceGitRepository `tfsdk:"git_repository"`
}

type ContainerSourceGitRepository struct {
	URL                    types.String `tfsdk:"url"`
	Branch                 types.String `tfsdk:"branch"`
	RootPath               types.String `tfsdk:"root_path"`
	GitTokenId             types.String `tfsdk:"git_token_id"`
	Provider               types.String `tfsdk:"provider"`
	DockerfilePath         types.String `tfsdk:"dockerfile_path"`
	DockerTargetBuildStage types.String `tfsdk:"docker_target_build_stage"`
}

func containerSourceResourceSchemaAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: containerSourceDescription + containerSourceWriteNote,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"git_repository": schema.SingleNestedAttribute{
				MarkdownDescription: containerSourceGitRepositoryDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: gitRepositoryURLDescription,
						Required:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: containerSourceBranchDescription + gitBranchRemovalNote,
						Optional:            true,
						Computed:            true,
						// Documented exception to the config-is-source-of-truth rule: the default
						// branch depends on the repository, so there is no static Default.
						PlanModifiers: []planmodifier.String{
							UseStateUnlessRepositoryChanges(),
						},
					},
					"root_path": schema.StringAttribute{
						MarkdownDescription: descriptions.NewStringDefaultDescription(gitRepositoryRootPathDescription+containerGitSourceRootPathWriteNote, containerGitSourceRootPathDefault),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(containerGitSourceRootPathDefault),
						// The read maps an API "" to "/", so "" in the configuration would never match the state.
						Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
					},
					"git_token_id": schema.StringAttribute{
						MarkdownDescription: gitRepositoryTokenIDDescription,
						Optional:            true,
						Validators:          []validator.String{validators.NewUUIDValidator()},
					},
					"provider": schema.StringAttribute{
						MarkdownDescription: gitRepositoryProviderDescription + gitRepositoryProviderConfigNote,
						Optional:            true,
						Validators: []validator.String{
							validators.NewStringEnumValidator(gitProviderValues),
						},
					},
					"dockerfile_path": schema.StringAttribute{
						MarkdownDescription: dockerfilePathDescription,
						Required:            true,
					},
					"docker_target_build_stage": schema.StringAttribute{
						MarkdownDescription: dockerTargetBuildStageDescription,
						Optional:            true,
					},
				},
			},
		},
	}
}

func containerSourceDataSourceSchemaAttribute() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		MarkdownDescription: containerSourceDescription,
		Computed:            true,
		Attributes: map[string]datasourceschema.Attribute{
			"git_repository": datasourceschema.SingleNestedAttribute{
				MarkdownDescription: containerSourceGitRepositoryDescription,
				Computed:            true,
				Attributes: map[string]datasourceschema.Attribute{
					"url":                       datasourceschema.StringAttribute{MarkdownDescription: gitRepositoryURLDescription, Computed: true},
					"branch":                    datasourceschema.StringAttribute{MarkdownDescription: containerSourceBranchDescription, Computed: true},
					"root_path":                 datasourceschema.StringAttribute{MarkdownDescription: gitRepositoryRootPathDescription, Computed: true},
					"git_token_id":              datasourceschema.StringAttribute{MarkdownDescription: gitRepositoryTokenIDDescription, Computed: true},
					"provider":                  datasourceschema.StringAttribute{MarkdownDescription: gitRepositoryProviderDescription, Computed: true},
					"dockerfile_path":           datasourceschema.StringAttribute{MarkdownDescription: dockerfilePathDescription, Computed: true},
					"docker_target_build_stage": datasourceschema.StringAttribute{MarkdownDescription: dockerTargetBuildStageDescription, Computed: true},
				},
			},
		},
	}
}

func (s *ContainerSource) gitRepository() *ContainerSourceGitRepository {
	if s == nil {
		return nil
	}
	return s.GitRepository
}

func (g *ContainerSourceGitRepository) toDomain() *docker.Docker {
	if g == nil {
		return nil
	}

	return &docker.Docker{
		GitRepository: git_repository.GitRepository{
			Url:        ToString(g.URL),
			Branch:     ToStringPointer(g.Branch),
			RootPath:   ToStringPointer(g.RootPath),
			GitTokenId: ToStringPointer(g.GitTokenId),
			Provider:   ToStringPointer(g.Provider),
		},
		DockerFilePath:         ToStringPointer(g.DockerfilePath),
		DockerTargetBuildStage: ToStringPointer(g.DockerTargetBuildStage),
	}
}

// containerSourceFromDomain returns the source to store for a container read from the API, nil for
// a container that runs a registry image. prior is the plan on apply and the state on refresh.
func containerSourceFromDomain(prior *ContainerSource, gitSource *docker.Docker) *ContainerSource {
	if gitSource == nil {
		return nil
	}

	var priorProvider *string
	priorGitTokenID := types.StringNull()
	priorTargetStage := types.StringNull()
	if priorGit := prior.gitRepository(); priorGit != nil {
		priorProvider = ToStringPointer(priorGit.Provider)
		priorGitTokenID = priorGit.GitTokenId
		priorTargetStage = priorGit.DockerTargetBuildStage
	}

	repository := gitSource.GitRepository
	return &ContainerSource{
		GitRepository: &ContainerSourceGitRepository{
			URL:                    FromString(repository.Url),
			Branch:                 FromStringPointer(repository.Branch),
			RootPath:               containerGitSourceRootPathFromAPI(repository.RootPath),
			GitTokenId:             optionalStringFromAPI(priorGitTokenID, repository.GitTokenId),
			Provider:               FromStringPointer(gitProviderFromDomain(repository.Url, repository.Provider, priorProvider)),
			DockerfilePath:         FromStringPointer(gitSource.DockerFilePath),
			DockerTargetBuildStage: optionalStringFromAPI(priorTargetStage, gitSource.DockerTargetBuildStage),
		},
	}
}

// containerGitSourceRootPathFromAPI reads an API "" or none as "/", the schema default: the engine
// builds both alike, as for jobs (jobRootPathFromAPI).
func containerGitSourceRootPathFromAPI(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringValue(containerGitSourceRootPathDefault)
	}
	return FromString(*v)
}

// containerSourceForDataSource reports the source as the API returns it.
func containerSourceForDataSource(gitSource *docker.Docker) *ContainerSource {
	if gitSource == nil {
		return nil
	}

	repository := gitSource.GitRepository
	return &ContainerSource{
		GitRepository: &ContainerSourceGitRepository{
			URL:                    FromString(repository.Url),
			Branch:                 FromStringPointer(repository.Branch),
			RootPath:               FromStringPointer(repository.RootPath),
			GitTokenId:             FromStringPointer(repository.GitTokenId),
			Provider:               FromStringPointer(repository.Provider),
			DockerfilePath:         FromStringPointer(gitSource.DockerFilePath),
			DockerTargetBuildStage: FromStringPointer(gitSource.DockerTargetBuildStage),
		},
	}
}

// attributeGetter is satisfied by tfsdk.Config, tfsdk.Plan and tfsdk.State.
type attributeGetter interface {
	GetAttribute(ctx context.Context, p path.Path, target interface{}) diag.Diagnostics
}

// containerSourceKind is what a container configuration, plan or state runs.
type containerSourceKind int

const (
	containerRunsImage containerSourceKind = iota
	containerSourceUnknown
	containerRunsGitRepository
)

// containerSourceKindAt reads source.git_repository of a container configuration, plan or state.
func containerSourceKindAt(ctx context.Context, data attributeGetter) (containerSourceKind, diag.Diagnostics) {
	var source types.Object
	diags := data.GetAttribute(ctx, path.Root("source"), &source)
	switch {
	case diags.HasError() || source.IsNull():
		return containerRunsImage, diags
	case source.IsUnknown():
		return containerSourceUnknown, diags
	}

	gitRepository, ok := source.Attributes()["git_repository"].(types.Object)
	switch {
	case !ok || gitRepository.IsNull():
		return containerRunsImage, diags
	case gitRepository.IsUnknown():
		return containerSourceUnknown, diags
	default:
		return containerRunsGitRepository, diags
	}
}

// gitContainerImageModifier keeps the image in state while the container keeps its Git repository
// source: only the API sets it there, to the last built image, so a build between two applies is no
// diff. On a switch to a Git repository source it stays unknown.
type gitContainerImageModifier struct{}

func (m gitContainerImageModifier) Description(_ context.Context) string {
	return "Keeps the image recorded in state while the container keeps its Git repository source."
}

func (m gitContainerImageModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m gitContainerImageModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	planned, diags := containerSourceKindAt(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	recorded, diags := containerSourceKindAt(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || planned != containerRunsGitRepository || recorded != containerRunsGitRepository {
		return
	}
	resp.PlanValue = req.StateValue
}

// GitContainerImage returns the plan modifier for registry_id, image_name and tag on qovery_container.
func GitContainerImage() planmodifier.String {
	return gitContainerImageModifier{}
}

// containerAutoDeployDefaultModifier plans the default of an omitted auto_deploy: true for a container
// that runs a registry image, false for one with a Git repository source, where the API rejects true.
type containerAutoDeployDefaultModifier struct{}

func (m containerAutoDeployDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Defaults to %t, or to false on a container with a Git repository source.", serviceAutoDeployDefault)
}

func (m containerAutoDeployDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m containerAutoDeployDefaultModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	planned, diags := containerSourceKindAt(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	switch planned {
	case containerSourceUnknown:
		resp.PlanValue = types.BoolUnknown()
	case containerRunsImage:
		resp.PlanValue = types.BoolValue(serviceAutoDeployDefault)
	case containerRunsGitRepository:
		resp.PlanValue = types.BoolValue(false)
	}
}

// ContainerAutoDeployDefault returns the plan modifier for auto_deploy on qovery_container.
func ContainerAutoDeployDefault() planmodifier.Bool {
	return containerAutoDeployDefaultModifier{}
}

// containerSourceConfig holds the configuration values that decide what a container runs.
type containerSourceConfig struct {
	Source               types.Object
	RegistryID           types.String
	ImageName            types.String
	Tag                  types.String
	BuildSettings        types.Object
	AutoDeploy           types.Bool
	AutoPreview          types.Bool
	AdvancedSettingsJson types.String
}

// validate checks that the container runs either a registry image or a Git repository source.
// Unknown values count as set.
func (c containerSourceConfig) validate() diag.Diagnostics {
	var diags diag.Diagnostics
	if c.Source.IsUnknown() {
		return diags
	}

	imageFields := []struct {
		name  string
		value types.String
	}{
		{"registry_id", c.RegistryID},
		{"image_name", c.ImageName},
		{"tag", c.Tag},
	}
	var set, missing []string
	for _, field := range imageFields {
		if field.value.IsNull() {
			missing = append(missing, field.name)
		} else {
			set = append(set, field.name)
		}
	}

	if !c.Source.IsNull() {
		if len(set) > 0 {
			diags.AddAttributeError(path.Root(set[0]), "Conflicting container image",
				fmt.Sprintf("Set either registry_id, image_name and tag, or source.git_repository, not both. "+
					"On a container with source.git_repository, Qovery reports the image it built in registry_id, image_name and tag: remove %s from the configuration.",
					strings.Join(set, ", ")))
		}
		if c.AutoDeploy.ValueBool() {
			diags.AddAttributeError(path.Root("auto_deploy"), "auto_deploy is not supported with a Git repository source",
				"Qovery does not redeploy a container with source.git_repository on new commits. Remove auto_deploy or set it to false.")
		}
		if c.AutoPreview.ValueBool() {
			diags.AddAttributeError(path.Root("auto_preview"), "auto_preview is not supported with a Git repository source",
				"Qovery does not create preview environments for a container with source.git_repository. Remove auto_preview or set it to false.")
		}
		diags.Append(validateBuildSettingsConflict(c.BuildSettings, c.AdvancedSettingsJson)...)
		return diags
	}

	if !c.BuildSettings.IsNull() {
		diags.AddAttributeError(path.Root("build_settings"), "build_settings needs a Git repository source",
			"Qovery builds no image for a container that runs a registry image. Remove build_settings, or set source.git_repository.")
	}

	switch len(missing) {
	case 0:
	case len(imageFields):
		diags.AddError("Missing container image",
			"Set registry_id, image_name and tag to run an image from a container registry, or source.git_repository to build the image from a Git repository.")
	default:
		diags.AddAttributeError(path.Root(missing[0]), "Incomplete container image",
			fmt.Sprintf("registry_id, image_name and tag go together: set %s too.", strings.Join(missing, ", ")))
	}

	return diags
}

// containerSourceValidator rejects a container configuration that runs both a registry image and a
// Git repository source, or neither, and the attributes the API rejects on either kind.
type containerSourceValidator struct{}

func (v containerSourceValidator) Description(_ context.Context) string {
	return "Checks that the container runs either a registry image or a Git repository source."
}

func (v containerSourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v containerSourceValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg containerSourceConfig
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("source"), &cfg.Source)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("registry_id"), &cfg.RegistryID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("image_name"), &cfg.ImageName)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("tag"), &cfg.Tag)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("build_settings"), &cfg.BuildSettings)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("auto_deploy"), &cfg.AutoDeploy)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("auto_preview"), &cfg.AutoPreview)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("advanced_settings_json"), &cfg.AdvancedSettingsJson)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(cfg.validate()...)
}
