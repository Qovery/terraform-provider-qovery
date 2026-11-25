//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

func TestContainerSourceConfigValidate(t *testing.T) {
	t.Parallel()

	gitSource := types.ObjectValueMust(map[string]attr.Type{"git_repository": types.StringType}, map[string]attr.Value{"git_repository": types.StringValue("set")})
	buildSettings := buildSettingsFromQovery(&qovery.BuildSettings{})

	image := func() containerSourceConfig {
		return containerSourceConfig{
			Source:        types.ObjectNull(map[string]attr.Type{}),
			RegistryID:    types.StringValue(uuid.NewString()),
			ImageName:     types.StringValue("nginx"),
			Tag:           types.StringValue("1.27"),
			BuildSettings: types.ObjectNull(buildSettingsAttrTypes),
		}
	}
	git := func() containerSourceConfig {
		cfg := image()
		cfg.Source = gitSource
		cfg.RegistryID = types.StringNull()
		cfg.ImageName = types.StringNull()
		cfg.Tag = types.StringNull()
		return cfg
	}

	testCases := []struct {
		TestName      string
		Config        func() containerSourceConfig
		ExpectedError string
	}{
		{TestName: "image", Config: image},
		{TestName: "git_repository_source", Config: git},
		{
			TestName: "git_repository_source_with_build_settings",
			Config: func() containerSourceConfig {
				cfg := git()
				cfg.BuildSettings = buildSettings
				return cfg
			},
		},
		{
			TestName: "unknown_image_counts_as_set",
			Config: func() containerSourceConfig {
				cfg := image()
				cfg.Tag = types.StringUnknown()
				return cfg
			},
		},
		{
			TestName: "unknown_source_is_not_checked",
			Config: func() containerSourceConfig {
				cfg := image()
				cfg.Source = types.ObjectUnknown(map[string]attr.Type{})
				return cfg
			},
		},
		{
			TestName: "both_image_and_git_repository_source",
			Config: func() containerSourceConfig {
				cfg := git()
				cfg.Tag = types.StringValue("1.27")
				return cfg
			},
			ExpectedError: "Conflicting container image",
		},
		{
			TestName: "neither_image_nor_git_repository_source",
			Config: func() containerSourceConfig {
				cfg := git()
				cfg.Source = types.ObjectNull(map[string]attr.Type{})
				return cfg
			},
			ExpectedError: "Missing container image",
		},
		{
			TestName: "incomplete_image",
			Config: func() containerSourceConfig {
				cfg := image()
				cfg.Tag = types.StringNull()
				return cfg
			},
			ExpectedError: "Incomplete container image",
		},
		{
			TestName: "build_settings_on_image",
			Config: func() containerSourceConfig {
				cfg := image()
				cfg.BuildSettings = buildSettings
				return cfg
			},
			ExpectedError: "build_settings needs a Git repository source",
		},
		{
			TestName: "build_settings_with_build_keys_in_advanced_settings",
			Config: func() containerSourceConfig {
				cfg := git()
				cfg.BuildSettings = buildSettings
				cfg.AdvancedSettingsJson = types.StringValue(`{"build.timeout_max_sec": 600}`)
				return cfg
			},
			ExpectedError: "Conflicting build configuration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			diags := tc.Config().validate()
			if tc.ExpectedError == "" {
				assert.False(t, diags.HasError(), "%v", diags)
				return
			}
			require.True(t, diags.HasError())
			assert.Equal(t, tc.ExpectedError, diags.Errors()[0].Summary())
		})
	}
}

func TestContainerToUpsertRepositoryRequest_Source(t *testing.T) {
	t.Parallel()

	imageContainer := nullPlanFor[Container](t, &containerResource{})
	imageContainer.Healthchecks = &HealthChecks{}
	imageContainer.Name = types.StringValue("api")
	imageContainer.RegistryID = types.StringValue("2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10")
	imageContainer.ImageName = types.StringValue("qovery/nginx")
	imageContainer.Tag = types.StringValue("1.27")

	gitContainer := nullPlanFor[Container](t, &containerResource{})
	gitContainer.Healthchecks = &HealthChecks{}
	gitContainer.Name = types.StringValue("api")
	gitContainer.RegistryID = types.StringValue("7c3e1a90-5b2d-4e8f-a1c6-0d9b8e7f6a54")
	gitContainer.ImageName = types.StringValue("z6f0c8a52-built")
	gitContainer.Tag = types.StringValue("4f2a9c1")
	gitContainer.Source = &ContainerSource{GitRepository: &ContainerSourceGitRepository{
		URL:                    types.StringValue("https://github.com/acme/api.git"),
		Branch:                 types.StringUnknown(),
		RootPath:               types.StringValue("/"),
		GitTokenId:             types.StringNull(),
		Provider:               types.StringNull(),
		DockerfilePath:         types.StringValue("Dockerfile"),
		DockerTargetBuildStage: types.StringNull(),
	}}
	gitContainer.BuildSettings = buildSettingsFromQovery(&qovery.BuildSettings{TimeoutMaxSec: new(int32(900))})

	t.Run("image_container", func(t *testing.T) {
		t.Parallel()
		request := imageContainer.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, &imageContainer)
		assert.Equal(t, "qovery/nginx", request.ImageName)
		assert.False(t, request.ImageAsSource)
		assert.Nil(t, request.GitSource)
		assert.Nil(t, request.BuildSettings)
	})

	t.Run("switch_to_git_repository_source", func(t *testing.T) {
		t.Parallel()
		request := gitContainer.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, &imageContainer)
		require.NoError(t, request.Validate())
		require.NotNil(t, request.GitSource)
		assert.Equal(t, "https://github.com/acme/api.git", request.GitSource.GitRepository.Url)
		assert.Nil(t, request.GitSource.GitRepository.Branch)
		assert.Equal(t, new("Dockerfile"), request.GitSource.DockerFilePath)
		assert.Empty(t, request.RegistryID)
		assert.Empty(t, request.ImageName)
		assert.Empty(t, request.Tag)
		require.NotNil(t, request.BuildSettings)
		assert.Equal(t, int32(900), request.BuildSettings.GetTimeoutMaxSec())
	})

	t.Run("switch_back_to_image", func(t *testing.T) {
		t.Parallel()
		request := imageContainer.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, &gitContainer)
		require.NoError(t, request.Validate())
		assert.True(t, request.ImageAsSource)
		assert.Equal(t, "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10", request.RegistryID)
		assert.Nil(t, request.GitSource)
		assert.Nil(t, request.BuildSettings)
	})

	t.Run("create_image_container", func(t *testing.T) {
		t.Parallel()
		request := imageContainer.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, nil)
		assert.False(t, request.ImageAsSource)
	})

	t.Run("create_git_container", func(t *testing.T) {
		t.Parallel()
		created := gitContainer
		created.RegistryID = types.StringUnknown()
		created.ImageName = types.StringUnknown()
		created.Tag = types.StringUnknown()
		created.AutoDeploy = types.BoolValue(false)
		created.AutoPreview = types.BoolValue(false)

		request := created.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, nil)
		require.NoError(t, request.Validate())
		require.NotNil(t, request.GitSource)
		assert.Equal(t, "https://github.com/acme/api.git", request.GitSource.GitRepository.Url)
		assert.Empty(t, request.RegistryID)
		assert.Empty(t, request.ImageName)
		assert.Empty(t, request.Tag)
		assert.False(t, request.ImageAsSource)
		require.NotNil(t, request.BuildSettings)
		assert.Equal(t, int32(900), request.BuildSettings.GetTimeoutMaxSec())
		assert.Equal(t, new(false), request.AutoDeploy.Get())
		assert.Equal(t, new(false), request.AutoPreview)
	})

	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprintf("update_git_container_sends_auto_flags_%t", enabled), func(t *testing.T) {
			t.Parallel()
			updated := gitContainer
			updated.AutoDeploy = types.BoolValue(enabled)
			updated.AutoPreview = types.BoolValue(enabled)

			request := updated.toUpsertRepositoryRequest(client.CustomDomainsDiff{}, &gitContainer)
			require.NoError(t, request.Validate())
			require.NotNil(t, request.GitSource)
			assert.Equal(t, new(enabled), request.AutoDeploy.Get())
			assert.Equal(t, new(enabled), request.AutoPreview)
		})
	}
}

func TestConvertDomainContainerToContainer_Source(t *testing.T) {
	t.Parallel()

	newContainer := func(gitSource *docker.Docker, buildSettings *qovery.BuildSettings) *container.Container {
		return &container.Container{
			ID:            uuid.New(),
			EnvironmentID: uuid.New(),
			RegistryID:    uuid.New(),
			Name:          "api",
			ImageName:     "z6f0c8a52-built",
			Tag:           "4f2a9c1",
			GitSource:     gitSource,
			BuildSettings: buildSettings,
		}
	}
	gitSource := &docker.Docker{
		GitRepository: git_repository.GitRepository{
			Url:        "https://github.com/acme/api.git",
			Branch:     new("main"),
			RootPath:   new("/"),
			GitTokenId: new("9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d"),
			Provider:   new("GITHUB"),
		},
		DockerFilePath: new("Dockerfile"),
	}
	prior := nullPlanFor[Container](t, &containerResource{})

	t.Run("image_container_has_no_source", func(t *testing.T) {
		t.Parallel()
		state := convertDomainContainerToContainer(context.Background(), prior, newContainer(nil, nil))
		assert.Nil(t, state.Source)
		assert.True(t, state.BuildSettings.IsNull())
	})

	t.Run("git_container_imported", func(t *testing.T) {
		t.Parallel()
		state := convertDomainContainerToContainer(context.Background(), prior, newContainer(gitSource, &qovery.BuildSettings{}))
		require.NotNil(t, state.Source)
		git := state.Source.GitRepository
		assert.Equal(t, "https://github.com/acme/api.git", git.URL.ValueString())
		assert.Equal(t, "main", git.Branch.ValueString())
		assert.Equal(t, "9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d", git.GitTokenId.ValueString())
		assert.True(t, git.Provider.IsNull(), "a provider detected from the URL is not stored")
		assert.True(t, git.DockerTargetBuildStage.IsNull())
		assert.Equal(t, "4f2a9c1", state.Tag.ValueString())
		assert.True(t, state.BuildSettings.IsNull(), "build_settings is only tracked once set")
	})

	t.Run("root_path_unset_or_empty_reads_as_slash", func(t *testing.T) {
		t.Parallel()
		for _, rootPath := range []*string{nil, new("")} {
			source := *gitSource
			source.GitRepository.RootPath = rootPath
			state := convertDomainContainerToContainer(context.Background(), prior, newContainer(&source, nil))
			require.NotNil(t, state.Source)
			assert.Equal(t, types.StringValue("/"), state.Source.GitRepository.RootPath)
		}
	})

	t.Run("data_source_reports_provider", func(t *testing.T) {
		t.Parallel()
		source := containerSourceForDataSource(gitSource)
		require.NotNil(t, source)
		assert.Equal(t, "GITHUB", source.GitRepository.Provider.ValueString())
	})
}

func TestConvertUpdatedContainerToContainer_Image(t *testing.T) {
	t.Parallel()

	gitSource := func() *ContainerSource {
		return &ContainerSource{GitRepository: &ContainerSourceGitRepository{
			URL:                    types.StringValue(testGitURL),
			Branch:                 types.StringValue("main"),
			RootPath:               types.StringValue("/"),
			GitTokenId:             types.StringNull(),
			Provider:               types.StringNull(),
			DockerfilePath:         types.StringValue("Dockerfile"),
			DockerTargetBuildStage: types.StringNull(),
		}}
	}
	withImage := func(source *ContainerSource, registryID, imageName, tag types.String) Container {
		cont := nullPlanFor[Container](t, &containerResource{})
		cont.Source = source
		cont.RegistryID = registryID
		cont.ImageName = imageName
		cont.Tag = tag
		return cont
	}
	builtImage := func(tag string) *container.Container {
		return &container.Container{
			ID:            uuid.MustParse(testContainerID),
			EnvironmentID: uuid.MustParse(testEnvironmentID),
			RegistryID:    uuid.MustParse(testBuildRegistryID),
			Name:          "api",
			ImageName:     "z6f0c8a52-built",
			Tag:           tag,
			GitSource: &docker.Docker{
				GitRepository:  git_repository.GitRepository{Url: testGitURL, Branch: new("main"), RootPath: new("/")},
				DockerFilePath: new("Dockerfile"),
			},
		}
	}
	builtValues := func(tag string) (types.String, types.String, types.String) {
		return types.StringValue(testBuildRegistryID), types.StringValue("z6f0c8a52-built"), types.StringValue(tag)
	}

	t.Run("git_container_keeps_the_planned_image_when_a_build_finished_after_the_plan", func(t *testing.T) {
		t.Parallel()
		registryID, imageName, tag := builtValues("build-1")
		prior := withImage(gitSource(), registryID, imageName, tag)
		plan := withImage(gitSource(), registryID, imageName, tag)

		state := convertUpdatedContainerToContainer(context.Background(), plan, prior, builtImage("build-2"))

		assert.Equal(t, "build-1", state.Tag.ValueString())
		assert.Equal(t, testBuildRegistryID, state.RegistryID.ValueString())
		assert.Equal(t, "z6f0c8a52-built", state.ImageName.ValueString())
	})

	t.Run("switch_back_to_image_stores_the_api_image", func(t *testing.T) {
		t.Parallel()
		registryID, imageName, tag := builtValues("build-1")
		prior := withImage(gitSource(), registryID, imageName, tag)
		plan := withImage(nil, types.StringValue(testRegistryID), types.StringValue("qovery/nginx"), types.StringValue("1.27"))
		running := &container.Container{
			ID:            uuid.MustParse(testContainerID),
			EnvironmentID: uuid.MustParse(testEnvironmentID),
			RegistryID:    uuid.MustParse(testRegistryID),
			Name:          "api",
			ImageName:     "qovery/nginx",
			Tag:           "1.28",
		}

		state := convertUpdatedContainerToContainer(context.Background(), plan, prior, running)

		assert.Equal(t, "1.28", state.Tag.ValueString())
		assert.Nil(t, state.Source)
	})

	t.Run("switch_to_git_stores_the_api_image", func(t *testing.T) {
		t.Parallel()
		prior := withImage(nil, types.StringValue(testRegistryID), types.StringValue("qovery/nginx"), types.StringValue("1.27"))
		plan := withImage(gitSource(), types.StringUnknown(), types.StringUnknown(), types.StringUnknown())

		state := convertUpdatedContainerToContainer(context.Background(), plan, prior, builtImage("build-2"))

		assert.Equal(t, "build-2", state.Tag.ValueString())
		assert.Equal(t, testBuildRegistryID, state.RegistryID.ValueString())
	})
}

const (
	testContainerID     = "6f0c8a52-3e0f-4c43-a1a8-5d6d1c0e9b21"
	testEnvironmentID   = "0e8f9b0c-7d0a-4f5e-9c1e-3a2b1c0d9e8f"
	testRegistryID      = "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10"
	testBuildRegistryID = "7c3e1a90-5b2d-4e8f-a1c6-0d9b8e7f6a54"
	testGitURL          = "https://github.com/acme/api.git"
)

func TestContainerPlan_ImageContainerUnchanged_NoDiff(t *testing.T) {
	t.Parallel()

	prior := testContainerValue(t, testImageContainerState())
	resp := planTestContainer(t, prior, testContainerValue(t, testImageContainerConfig()))

	requireNoPlanErrors(t, resp)
	assert.Empty(t, resp.RequiresReplace)
	assert.True(t, prior.Equal(plannedTestContainer(t, resp)), "an image container config planned a change")
}

func TestContainerPlan_ImageContainer_AutoDeployDefaultsToTrue(t *testing.T) {
	t.Parallel()

	s := testContainerSchema(t)
	created := planTestContainer(t, tftypes.NewValue(s.Type().TerraformType(context.Background()), nil), testContainerValue(t, testImageContainerConfig()))
	requireNoPlanErrors(t, created)
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, true), testValueAttributes(t, plannedTestContainer(t, created))["auto_deploy"])

	drifted := testImageContainerState()
	drifted["auto_deploy"] = tftypes.NewValue(tftypes.Bool, false)
	updated := planTestContainer(t, testContainerValue(t, drifted), testContainerValue(t, testImageContainerConfig()))
	requireNoPlanErrors(t, updated)
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, true), testValueAttributes(t, plannedTestContainer(t, updated))["auto_deploy"])
}

func TestContainerPlan_SwitchImageToGit_IsInPlaceUpdate(t *testing.T) {
	t.Parallel()

	config := testContainerValue(t, map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, testEnvironmentID),
		"name":           tftypes.NewValue(tftypes.String, "api"),
		"source":         testGitSourceValue(nil),
	})
	resp := planTestContainer(t, testContainerValue(t, testImageContainerState()), config)

	requireNoPlanErrors(t, resp)
	assert.Empty(t, resp.RequiresReplace)
	planned := testValueAttributes(t, plannedTestContainer(t, resp))
	assert.Equal(t, tftypes.NewValue(tftypes.String, testContainerID), planned["id"])
	for _, name := range []string{"registry_id", "image_name", "tag"} {
		assert.False(t, planned[name].IsKnown(), "%s must be unknown on a switch", name)
	}
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_deploy"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_preview"])
}

func TestContainerPlan_GitContainerAfterNewBuild_NoDiff(t *testing.T) {
	t.Parallel()

	prior := testContainerValue(t, testGitContainerState())
	resp := planTestContainer(t, prior, testContainerValue(t, testGitContainerConfig(nil)))

	requireNoPlanErrors(t, resp)
	assert.Empty(t, resp.RequiresReplace)
	assert.True(t, prior.Equal(plannedTestContainer(t, resp)), "a new build planned a change")
}

func TestContainerPlan_GitContainerUpdate_KeepsBuiltImage(t *testing.T) {
	t.Parallel()

	config := testGitContainerConfig(map[string]tftypes.Value{"cpu": tftypes.NewValue(tftypes.Number, big.NewFloat(1000))})
	resp := planTestContainer(t, testContainerValue(t, testGitContainerState()), testContainerValue(t, config))

	requireNoPlanErrors(t, resp)
	assert.Empty(t, resp.RequiresReplace)
	planned := testValueAttributes(t, plannedTestContainer(t, resp))
	assert.Equal(t, tftypes.NewValue(tftypes.String, testBuildRegistryID), planned["registry_id"])
	assert.Equal(t, tftypes.NewValue(tftypes.String, "z6f0c8a52-built"), planned["image_name"])
	assert.Equal(t, tftypes.NewValue(tftypes.String, "build-2"), planned["tag"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_deploy"])
}

func TestContainerPlan_SwitchGitToImage_IsInPlaceUpdate(t *testing.T) {
	t.Parallel()

	resp := planTestContainer(t, testContainerValue(t, testGitContainerState()), testContainerValue(t, testImageContainerConfig()))

	requireNoPlanErrors(t, resp)
	assert.Empty(t, resp.RequiresReplace)
	planned := testValueAttributes(t, plannedTestContainer(t, resp))
	assert.Equal(t, tftypes.NewValue(tftypes.String, testContainerID), planned["id"])
	assert.True(t, planned["source"].IsNull())
	assert.Equal(t, tftypes.NewValue(tftypes.String, testRegistryID), planned["registry_id"])
	assert.Equal(t, tftypes.NewValue(tftypes.String, "1.27"), planned["tag"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, true), planned["auto_deploy"])
}

func TestContainerPlan_EnvironmentChange_StillReplaces(t *testing.T) {
	t.Parallel()

	config := testImageContainerConfig()
	config["environment_id"] = tftypes.NewValue(tftypes.String, uuid.NewString())
	resp := planTestContainer(t, testContainerValue(t, testImageContainerState()), testContainerValue(t, config))

	requireNoPlanErrors(t, resp)
	require.Len(t, resp.RequiresReplace, 1)
	assert.True(t, resp.RequiresReplace[0].Equal(tftypes.NewAttributePath().WithAttributeName("environment_id")))
}

func TestContainerPlan_CreateWithGitSource_IsAccepted(t *testing.T) {
	t.Parallel()

	s := testContainerSchema(t)
	prior := tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)
	resp := planTestContainer(t, prior, testContainerValue(t, testGitContainerConfig(nil)))

	requireNoPlanErrors(t, resp)
	requireCreatePlanOfGitContainer(t, plannedTestContainer(t, resp))
}

// TestContainerPlan_ReplaceGitContainer_IsAccepted replays the two calls Terraform core makes for a
// replacement: the update plan that requires it, then the plan of the create from a null prior.
func TestContainerPlan_ReplaceGitContainer_IsAccepted(t *testing.T) {
	t.Parallel()

	config := testContainerValue(t, testGitContainerConfig(map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, uuid.NewString()),
	}))
	update := planTestContainer(t, testContainerValue(t, testGitContainerState()), config)
	requireNoPlanErrors(t, update)
	require.Len(t, update.RequiresReplace, 1)
	assert.True(t, update.RequiresReplace[0].Equal(tftypes.NewAttributePath().WithAttributeName("environment_id")))

	s := testContainerSchema(t)
	create := planTestContainer(t, tftypes.NewValue(s.Type().TerraformType(context.Background()), nil), config)
	requireNoPlanErrors(t, create)
	requireCreatePlanOfGitContainer(t, plannedTestContainer(t, create))
}

// TestContainerPlan_GitContainer_AutoFlags plans auto_deploy and auto_preview set to true on a Git
// container, then removed from the configuration: they go back to their default, false.
func TestContainerPlan_GitContainer_AutoFlags(t *testing.T) {
	t.Parallel()

	enabled := map[string]tftypes.Value{
		"auto_deploy":  tftypes.NewValue(tftypes.Bool, true),
		"auto_preview": tftypes.NewValue(tftypes.Bool, true),
	}
	s := testContainerSchema(t)
	created := planTestContainer(t, tftypes.NewValue(s.Type().TerraformType(context.Background()), nil), testContainerValue(t, testGitContainerConfig(enabled)))
	requireNoPlanErrors(t, created)
	planned := testValueAttributes(t, plannedTestContainer(t, created))
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, true), planned["auto_deploy"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, true), planned["auto_preview"])

	state := testGitContainerState()
	for name, value := range enabled {
		state[name] = value
	}
	removed := planTestContainer(t, testContainerValue(t, state), testContainerValue(t, testGitContainerConfig(nil)))
	requireNoPlanErrors(t, removed)
	planned = testValueAttributes(t, plannedTestContainer(t, removed))
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_deploy"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_preview"])
}

// requireCreatePlanOfGitContainer checks the create plan of testGitContainerConfig: the API sets
// the image and the built-in variables, and the auto flags take their Git container default, false.
func requireCreatePlanOfGitContainer(t *testing.T, plannedValue tftypes.Value) {
	t.Helper()
	planned := testValueAttributes(t, plannedValue)
	for _, name := range []string{"id", "registry_id", "image_name", "tag", "built_in_environment_variables"} {
		assert.False(t, planned[name].IsKnown(), "%s must be unknown on create", name)
	}
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_deploy"])
	assert.Equal(t, tftypes.NewValue(tftypes.Bool, false), planned["auto_preview"])
	assert.True(t, planned["build_settings"].IsNull())
}

// TestContainerCreateWithGitSource_ApplyThenPlan_NoDiff replays a create from a Git repository:
// until the first build, the API reports the cluster registry, the repository name and a
// placeholder tag as the image.
func TestContainerCreateWithGitSource_ApplyThenPlan_NoDiff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := testContainerSchema(t)

	config := testContainerValue(t, testGitContainerConfig(map[string]tftypes.Value{
		"healthchecks":   testEmptyHealthchecksValue(),
		"build_settings": testBuildSettingsValue(map[string]tftypes.Value{"timeout_max_sec": tftypes.NewValue(tftypes.Number, big.NewFloat(900))}),
	}))
	create := planTestContainer(t, tftypes.NewValue(s.Type().TerraformType(ctx), nil), config)
	requireNoPlanErrors(t, create)
	plannedValue := plannedTestContainer(t, create)

	var plan Container
	diags := tfsdk.Plan{Schema: s, Raw: plannedValue}.Get(ctx, &plan)
	require.False(t, diags.HasError(), diags)

	created := &container.Container{
		ID:                   uuid.MustParse(testContainerID),
		EnvironmentID:        uuid.MustParse(testEnvironmentID),
		RegistryID:           uuid.MustParse(testBuildRegistryID),
		Name:                 "api",
		IconUri:              containerIconURIDefault,
		ImageName:            "api",
		Tag:                  "pending-first-build",
		CPU:                  container.DefaultCPU,
		Memory:               container.DefaultMemory,
		MinRunningInstances:  container.MinMinRunningInstances,
		MaxRunningInstances:  container.DefaultMaxRunningInstances,
		InternalHost:         new("api.internal"),
		DeploymentStageID:    "8b6a2f4e-1c3d-4e5f-9a7b-2c1d0e9f8a7b",
		AdvancedSettingsJson: "{}",
		AutoDeploy:           new(false),
		GitSource: &docker.Docker{
			GitRepository:  git_repository.GitRepository{Url: testGitURL, Branch: new("main"), RootPath: new("/")},
			DockerFilePath: new("Dockerfile"),
		},
		BuildSettings: &qovery.BuildSettings{
			TimeoutMaxSec:        new(int32(900)),
			CpuMaxInMilli:        new(int32(buildSettingsDefaultCpuMaxInMilli)),
			RamMaxInGib:          new(int32(buildSettingsDefaultRamMaxInGib)),
			DisableBuildkitCache: new(buildSettingsDefaultDisableBuildkitCache),
			SkipGitSubmodules:    new(buildSettingsDefaultSkipGitSubmodules),
		},
	}
	state := tfsdk.State{Schema: s}
	diags = state.Set(ctx, convertDomainContainerToContainer(ctx, plan, created))
	require.False(t, diags.HasError(), diags)

	requireAppliedMatchesPlan(t, plannedValue, state.Raw)
	applied := testValueAttributes(t, state.Raw)
	assert.Equal(t, tftypes.NewValue(tftypes.String, testBuildRegistryID), applied["registry_id"])
	assert.Equal(t, tftypes.NewValue(tftypes.String, "api"), applied["image_name"])
	assert.Equal(t, tftypes.NewValue(tftypes.String, "pending-first-build"), applied["tag"])
	buildSettings := testValueAttributes(t, applied["build_settings"])
	assert.True(t, tftypes.NewValue(tftypes.Number, big.NewFloat(900)).Equal(buildSettings["timeout_max_sec"]), "timeout_max_sec: %s", buildSettings["timeout_max_sec"])

	next := planTestContainer(t, state.Raw, config)
	requireNoPlanErrors(t, next)
	assert.Empty(t, next.RequiresReplace)
	assert.True(t, state.Raw.Equal(plannedTestContainer(t, next)), "the plan after a create from a Git repository planned a change")
}

// requireAppliedMatchesPlan fails where Terraform reports "Provider produced inconsistent result
// after apply": the applied value differs from a known planned value.
func requireAppliedMatchesPlan(t *testing.T, planned, applied tftypes.Value) {
	t.Helper()
	diffs, err := planned.Diff(applied)
	require.NoError(t, err)
	for _, d := range diffs {
		if d.Value1 != nil && d.Value1.IsFullyKnown() {
			t.Errorf("%s: planned %s, applied %v", d.Path, d.Value1, d.Value2)
		}
	}
}

// testBuildSettingsValue builds a build_settings value from fields, every other setting null.
func testBuildSettingsValue(fields map[string]tftypes.Value) tftypes.Value {
	buildSettingsType := testContainerType().AttributeTypes["build_settings"].(tftypes.Object)
	settings := make(map[string]tftypes.Value, len(buildSettingsType.AttributeTypes))
	for name, settingType := range buildSettingsType.AttributeTypes {
		if value, ok := fields[name]; ok {
			settings[name] = value
		} else {
			settings[name] = tftypes.NewValue(settingType, nil)
		}
	}
	return tftypes.NewValue(buildSettingsType, settings)
}

func testEmptyHealthchecksValue() tftypes.Value {
	healthchecksType := testContainerType().AttributeTypes["healthchecks"].(tftypes.Object)
	probes := make(map[string]tftypes.Value, len(healthchecksType.AttributeTypes))
	for name, probeType := range healthchecksType.AttributeTypes {
		probes[name] = tftypes.NewValue(probeType, nil)
	}
	return tftypes.NewValue(healthchecksType, probes)
}

func TestContainerValidateConfig_RunsSourceValidator(t *testing.T) {
	t.Parallel()

	config := testGitContainerConfig(map[string]tftypes.Value{"tag": tftypes.NewValue(tftypes.String, "1.27")})
	value := testContainerValue(t, config)
	s := testContainerSchema(t)

	resp, err := providerserver.NewProtocol6(New("test")())().ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{
		TypeName: "qovery_container",
		Config:   testDynamicValue(t, s, value),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Diagnostics)
	assert.Equal(t, "Conflicting container image", resp.Diagnostics[0].Summary)
}

func TestContainerValidateConfig_GitSourceAcceptsAutoFlags(t *testing.T) {
	t.Parallel()

	config := testGitContainerConfig(map[string]tftypes.Value{
		"healthchecks": testEmptyHealthchecksValue(),
		"auto_deploy":  tftypes.NewValue(tftypes.Bool, true),
		"auto_preview": tftypes.NewValue(tftypes.Bool, true),
	})
	s := testContainerSchema(t)

	resp, err := providerserver.NewProtocol6(New("test")())().ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{
		TypeName: "qovery_container",
		Config:   testDynamicValue(t, s, testContainerValue(t, config)),
	})
	require.NoError(t, err)
	for _, d := range resp.Diagnostics {
		assert.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
}

func testImageContainerConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, testEnvironmentID),
		"name":           tftypes.NewValue(tftypes.String, "api"),
		"registry_id":    tftypes.NewValue(tftypes.String, testRegistryID),
		"image_name":     tftypes.NewValue(tftypes.String, "qovery/nginx"),
		"tag":            tftypes.NewValue(tftypes.String, "1.27"),
	}
}

// testImageContainerState is the state an image container has after a refresh, with the defaults
// of the attributes its configuration omits.
func testImageContainerState() map[string]tftypes.Value {
	state := testImageContainerConfig()
	for name, value := range testContainerComputedState() {
		state[name] = value
	}
	return state
}

func testGitContainerConfig(overrides map[string]tftypes.Value) map[string]tftypes.Value {
	config := map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, testEnvironmentID),
		"name":           tftypes.NewValue(tftypes.String, "api"),
		"source":         testGitSourceValue(nil),
	}
	for name, value := range overrides {
		config[name] = value
	}
	return config
}

// testGitContainerState is the state of a Git container refreshed after its second build.
func testGitContainerState() map[string]tftypes.Value {
	state := map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, testEnvironmentID),
		"name":           tftypes.NewValue(tftypes.String, "api"),
		"registry_id":    tftypes.NewValue(tftypes.String, testBuildRegistryID),
		"image_name":     tftypes.NewValue(tftypes.String, "z6f0c8a52-built"),
		"tag":            tftypes.NewValue(tftypes.String, "build-2"),
		"source": testGitSourceValue(map[string]tftypes.Value{
			"branch":    tftypes.NewValue(tftypes.String, "main"),
			"root_path": tftypes.NewValue(tftypes.String, "/"),
		}),
	}
	for name, value := range testContainerComputedState() {
		state[name] = value
	}
	state["auto_deploy"] = tftypes.NewValue(tftypes.Bool, false)
	return state
}

func testContainerComputedState() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, testContainerID),
		"icon_uri":               tftypes.NewValue(tftypes.String, containerIconURIDefault),
		"cpu":                    tftypes.NewValue(tftypes.Number, big.NewFloat(container.DefaultCPU)),
		"memory":                 tftypes.NewValue(tftypes.Number, big.NewFloat(container.DefaultMemory)),
		"ephemeral_storage":      tftypes.NewValue(tftypes.Number, big.NewFloat(float64(serviceEphemeralStorageDefault))),
		"min_running_instances":  tftypes.NewValue(tftypes.Number, big.NewFloat(container.MinMinRunningInstances)),
		"max_running_instances":  tftypes.NewValue(tftypes.Number, big.NewFloat(container.DefaultMaxRunningInstances)),
		"auto_preview":           tftypes.NewValue(tftypes.Bool, serviceAutoPreviewDefault),
		"auto_deploy":            tftypes.NewValue(tftypes.Bool, serviceAutoDeployDefault),
		"is_skipped":             tftypes.NewValue(tftypes.Bool, false),
		"deployment_stage_id":    tftypes.NewValue(tftypes.String, "8b6a2f4e-1c3d-4e5f-9a7b-2c1d0e9f8a7b"),
		"advanced_settings_json": tftypes.NewValue(tftypes.String, "{}"),
		"internal_host":          tftypes.NewValue(tftypes.String, "api.internal"),
	}
}

// testGitSourceValue builds the source of testGitURL with a Dockerfile at the root, fields added.
func testGitSourceValue(fields map[string]tftypes.Value) tftypes.Value {
	sourceType := testContainerType().AttributeTypes["source"].(tftypes.Object)
	gitType := sourceType.AttributeTypes["git_repository"].(tftypes.Object)

	git := map[string]tftypes.Value{
		"url":             tftypes.NewValue(tftypes.String, testGitURL),
		"dockerfile_path": tftypes.NewValue(tftypes.String, "Dockerfile"),
	}
	for name, value := range fields {
		git[name] = value
	}
	for name, attrType := range gitType.AttributeTypes {
		if _, ok := git[name]; !ok {
			git[name] = tftypes.NewValue(attrType, nil)
		}
	}

	return tftypes.NewValue(sourceType, map[string]tftypes.Value{
		"git_repository": tftypes.NewValue(gitType, git),
	})
}

func testContainerSchema(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	containerResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp.Schema
}

func testContainerType() tftypes.Object {
	var resp resource.SchemaResponse
	containerResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
}

// testContainerValue builds a qovery_container value from attrs, every other attribute null.
func testContainerValue(t *testing.T, attrs map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	containerType := testContainerType()
	values := make(map[string]tftypes.Value, len(containerType.AttributeTypes))
	for name, attrType := range containerType.AttributeTypes {
		if value, ok := attrs[name]; ok {
			values[name] = value
		} else {
			values[name] = tftypes.NewValue(attrType, nil)
		}
	}
	return tftypes.NewValue(containerType, values)
}

// testProposedNewState mimics the proposed new state Terraform sends: the configuration, where a
// computed attribute the configuration leaves null takes its prior value.
func testProposedNewState(attributes map[string]schema.Attribute, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() || !config.IsKnown() {
		return config
	}

	var configValues, priorValues map[string]tftypes.Value
	_ = config.As(&configValues)
	if !prior.IsNull() {
		_ = prior.As(&priorValues)
	}

	proposed := make(map[string]tftypes.Value, len(configValues))
	for name, value := range configValues {
		priorValue, hasPrior := priorValues[name]
		if !hasPrior {
			priorValue = tftypes.NewValue(value.Type(), nil)
		}
		nested, isNested := attributes[name].(schema.SingleNestedAttribute)
		switch {
		case value.IsNull() && attributes[name].IsComputed():
			proposed[name] = priorValue
		case isNested:
			proposed[name] = testProposedNewState(nested.Attributes, priorValue, value)
		default:
			proposed[name] = value
		}
	}
	return tftypes.NewValue(config.Type(), proposed)
}

func testDynamicValue(t *testing.T, s schema.Schema, value tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()
	dv, err := tfprotov6.NewDynamicValue(s.Type().TerraformType(context.Background()), value)
	require.NoError(t, err)
	return &dv
}

func planTestContainer(t *testing.T, prior, config tftypes.Value) *tfprotov6.PlanResourceChangeResponse {
	t.Helper()
	s := testContainerSchema(t)
	resp, err := providerserver.NewProtocol6(New("test")())().PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "qovery_container",
		PriorState:       testDynamicValue(t, s, prior),
		ProposedNewState: testDynamicValue(t, s, testProposedNewState(s.Attributes, prior, config)),
		Config:           testDynamicValue(t, s, config),
	})
	require.NoError(t, err)
	return resp
}

func plannedTestContainer(t *testing.T, resp *tfprotov6.PlanResourceChangeResponse) tftypes.Value {
	t.Helper()
	require.NotNil(t, resp.PlannedState)
	planned, err := resp.PlannedState.Unmarshal(testContainerType())
	require.NoError(t, err)
	return planned
}

func testValueAttributes(t *testing.T, value tftypes.Value) map[string]tftypes.Value {
	t.Helper()
	var attributes map[string]tftypes.Value
	require.NoError(t, value.As(&attributes))
	return attributes
}

func requireNoPlanErrors(t *testing.T, resp *tfprotov6.PlanResourceChangeResponse) {
	t.Helper()
	for _, d := range resp.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
}
