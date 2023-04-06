//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

const (
	containerImageName = "terraform-provider-tests-container"
	containerTag       = "1.0.0"
)

func TestAcc_Container(t *testing.T) {
	t.Parallel()
	testName := "container"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Update name
			{
				Config: testAccContainerDefaultConfigWithName(
					testName,
					fmt.Sprintf("%s-updated", testName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", fmt.Sprintf("%s-updated", testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithArguments(t *testing.T) {
	t.Parallel()
	testName := "container-with-arguments"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithArguments(
					testName,
					[]string{"arg1"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "arguments.0", "arg1"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Add argument
			{
				Config: testAccContainerDefaultConfigWithArguments(
					testName,
					[]string{"arg1", "arg2"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "arguments.0", "arg1"),
					resource.TestCheckResourceAttr("qovery_container.test", "arguments.1", "arg2"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Remove argument
			{
				Config: testAccContainerDefaultConfigWithArguments(
					testName,
					[]string{"arg2"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "arguments.0", "arg2"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithAutoPreview(t *testing.T) {
	t.Parallel()
	testName := "container-with-auto-preview"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithAutoPreview(
					testName,
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "true"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Update auto_preview
			{
				Config: testAccContainerDefaultConfigWithAutoPreview(
					testName,
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithResources(t *testing.T) {
	t.Parallel()
	testName := "container-with-resources"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithResources(
					testName,
					"1000",
					"1024",
					"2",
					"3",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "1000"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "1024"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "3"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Update auto_preview
			{
				Config: testAccContainerDefaultConfigWithResources(
					testName,
					"1500",
					"2048",
					"4",
					"6",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "1500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "2048"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "4"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "6"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithEnvironmentVariables(t *testing.T) {
	t.Parallel()
	testName := "container-with-environment-variables"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Update environment variable
			{
				Config: testAccContainerDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Add environment variable
			{
				Config: testAccContainerDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Remove environment variables
			{
				Config: testAccContainerDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithSecrets(t *testing.T) {
	t.Parallel()
	testName := "container-with-secrets"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Update secret
			{
				Config: testAccContainerDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Add secret
			{
				Config: testAccContainerDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Remove secret
			{
				Config: testAccContainerDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
		},
	})
}

//
//func TestAcc_ContainerWithCustomDomains(t *testing.T) {
//	t.Parallel()
//	testName := "container-with-custom-domains"
//	// NOTE: Run this test with stopped container unless we are in the main branch.
//	// Running it with a STOPPED state will make the test run way faster.
//	state := "STOPPED"
//	if isCIMainBranch() {
//		state = "RUNNING"
//	}
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccContainerDefaultConfigWithCustomDomains(
//					testName,
//					[]string{"toto.com"},
//					state,
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.internal_port", "8000"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.publicly_accessible", "true"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "custom_domains.*", map[string]string{
//						"domain": "toto.com",
//					}),
//					resource.TestCheckResourceAttrSet("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Update environment variable
//			{
//				Config: testAccContainerDefaultConfigWithCustomDomains(
//					testName,
//					[]string{"toto-updated.com"},
//					state,
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.internal_port", "8000"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.publicly_accessible", "true"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "custom_domains.*", map[string]string{
//						"domain": "toto-updated.com",
//					}),
//					resource.TestCheckResourceAttrSet("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Add environment variable
//			{
//				Config: testAccContainerDefaultConfigWithCustomDomains(
//					testName,
//					[]string{"toto.com", "tata.com"},
//					state,
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.internal_port", "8000"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.publicly_accessible", "true"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "custom_domains.*", map[string]string{
//						"domain": "toto.com",
//					}),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "custom_domains.*", map[string]string{
//						"domain": "tata.com",
//					}),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttrSet("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Remove environment variables
//			{
//				Config: testAccContainerDefaultConfigWithCustomDomains(
//					testName,
//					[]string{"tata.com"},
//					state,
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.internal_port", "8000"),
//					resource.TestCheckResourceAttr("qovery_container.test", "ports.0.publicly_accessible", "true"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "custom_domains.*", map[string]string{
//						"domain": "tata.com",
//					}),
//					resource.TestCheckResourceAttrSet("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Check Import
//			{
//				ResourceName:      "qovery_container.test",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}
//
//// Container should redeploy when environment env variables are updated.
//func TestAcc_ContainerRedeployOnEnvironmentUpdate(t *testing.T) {
//	t.Parallel()
//	testName := "container-redeploy"
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccContainerDefaultConfig(
//					testName,
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Update environment env variables
//			{
//				Config: testAccContainerDefaultConfigWithEnvironmentEnvVariables(
//					testName,
//					map[string]string{
//						"key1": "value1",
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
//						"key":   "key1",
//						"value": "value1",
//					}),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Update environment variables
//			{
//				Config: testAccContainerDefaultConfigWithEnvironmentEnvVariables(
//					testName,
//					map[string]string{
//						"key1": "value1-updated",
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryContainerExists("qovery_container.test"),
//					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.url", containerRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.branch", containerBranch),
//					resource.TestCheckResourceAttr("qovery_container.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_container.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_container.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
//					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
//						"key":   "key1",
//						"value": "value1-updated",
//					}),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
//					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^app-z`)),
//				),
//			},
//			// Check Import
//			{
//				ResourceName:      "qovery_container.test",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}

// TODO: uncomment after debugging why storage can't be updated
func TestAcc_ContainerWithStorage(t *testing.T) {
	t.Parallel()
	testName := "container-with-storage"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithStorage(
					testName,
					[]serviceStorage{
						{
							Type:       "FAST_SSD",
							Size:       1,
							MountPoint: "/data",
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "storage.*", map[string]string{
						"type":        "FAST_SSD",
						"size":        "1",
						"mount_point": "/data",
					}),

					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Add another storage
			{
				Config: testAccContainerDefaultConfigWithStorage(
					testName,
					[]serviceStorage{
						{
							Type:       "FAST_SSD",
							Size:       1,
							MountPoint: "/toto",
						},
						{
							Type:       "FAST_SSD",
							Size:       1,
							MountPoint: "/tata",
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "storage.*", map[string]string{
						"type":        "FAST_SSD",
						"size":        "1",
						"mount_point": "/toto",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "storage.*", map[string]string{
						"type":        "FAST_SSD",
						"size":        "1",
						"mount_point": "/tata",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			// Remove first storage
			{
				Config: testAccContainerDefaultConfigWithStorage(
					testName,
					[]serviceStorage{
						{
							Type:       "FAST_SSD",
							Size:       1,
							MountPoint: "/toto",
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "storage.*", map[string]string{
						"type":        "FAST_SSD",
						"size":        "1",
						"mount_point": "/toto",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithPorts(t *testing.T) {
	t.Parallel()
	testName := "container-with-ports"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithPorts(
					testName,
					[]servicePort{
						{
							InternalPort:       80,
							PubliclyAccessible: false,
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "ports.*", map[string]string{
						"internal_port":       "80",
						"publicly_accessible": "false",
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
				),
			},
			//// Add another port
			//{
			//	Config: testAccContainerDefaultConfigWithPorts(
			//		testName,
			//		[]servicePort{
			//			{
			//				InternalPort:       80,
			//				PubliclyAccessible: false,
			//			},
			//			{
			//				Name:               pointer.ToString("external port"),
			//				InternalPort:       81,
			//				ExternalPort:       int64ToPtr(443),
			//				PubliclyAccessible: true,
			//				Protocol:           pointer.ToString("HTTP"),
			//			},
			//		},
			//	),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		testAccQoveryProjectExists("qovery_project.test"),
			//		testAccQoveryEnvironmentExists("qovery_environment.test"),
			//		testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
			//		testAccQoveryContainerExists("qovery_container.test"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
			//		resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
			//		resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
			//		resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
			//		resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "ports.*", map[string]string{
			//			"internal_port":       "80",
			//			"publicly_accessible": "false",
			//		}),
			//		resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "ports.*", map[string]string{
			//			"name":                "external port",
			//			"internal_port":       "81",
			//			"external_port":       "443",
			//			"publicly_accessible": "true",
			//			"protocol":            "HTTP",
			//		}),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
			//		resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
			//			"key": regexp.MustCompile(`^QOVERY_`),
			//		}),
			//		resource.TestMatchResourceAttr("qovery_container.test", "external_host", regexp.MustCompile(`^z`)),
			//		resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
			//	),
			//},
			//// Remove first port
			//{
			//	Config: testAccContainerDefaultConfigWithPorts(
			//		testName,
			//		[]servicePort{
			//			{
			//				Name:               pointer.ToString("external port"),
			//				InternalPort:       81,
			//				ExternalPort:       int64ToPtr(443),
			//				PubliclyAccessible: true,
			//				Protocol:           pointer.ToString("HTTP"),
			//			},
			//		},
			//	),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		testAccQoveryProjectExists("qovery_project.test"),
			//		testAccQoveryEnvironmentExists("qovery_environment.test"),
			//		testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
			//		testAccQoveryContainerExists("qovery_container.test"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
			//		resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
			//		resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
			//		resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
			//		resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "storage.0"),
			//		resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "ports.*", map[string]string{
			//			"name":                "external port",
			//			"internal_port":       "81",
			//			"external_port":       "443",
			//			"publicly_accessible": "true",
			//			"protocol":            "HTTP",
			//		}),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
			//		resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
			//		resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
			//			"key": regexp.MustCompile(`^QOVERY_`),
			//		}),
			//		resource.TestMatchResourceAttr("qovery_container.test", "external_host", regexp.MustCompile(`^z`)),
			//		resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
			//	),
			//},
			// Check Import
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ContainerWithAdvSettings(t *testing.T) {
	t.Parallel()
	testName := "container-with-adv-settings"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerDefaultConfigWithAdvSettings(
					testName,
					containerAdvSettingsString(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.deployment.custom_domain_check_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.deployment.termination_grace_period_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.hpa.cpu.average_utilization_percent", "70"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.failure_threshold", "9"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.http_get.path", "/"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.initial_delay_seconds", "15"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.period_seconds", "5"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.success_threshold", "3"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.timeout_seconds", "3"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.type", "TCP"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.basic_auth_env_var", ""),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_headers", "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_methods", "GET, PUT, POST, DELETE, PATCH, OPTIONS"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_origin", "*"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.denylist_source_range", ""),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.enable_cors", "true"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.enable_sticky_session", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.keepalive_time_seconds", "1800"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.keepalive_timeout_seconds", "60"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_body_size_mb", "50"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_buffer_size_kb", "20"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_connect_timeout_seconds", "30"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_read_timeout_seconds", "30"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_send_timeout_seconds", "30"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.send_timeout_seconds", "30"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.whitelist_source_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.failure_threshold", "8"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.http_get.path", "/"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.initial_delay_seconds", "15"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.period_seconds", "5"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.success_threshold", "10"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.timeout_seconds", "10"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.type", "TCP"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.security.service_account_name", ""),
				),
			},
			// Update advanced settings
			{
				Config: testAccContainerDefaultConfigWithAdvSettings(
					testName,
					editedContainerAdvSettingsString(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "arguments.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_container.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_container.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_container.test", "internal_host", regexp.MustCompile(`^container-z`)),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.deployment.custom_domain_check_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.deployment.termination_grace_period_seconds", "240"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.hpa.cpu.average_utilization_percent", "80"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.failure_threshold", "12"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.http_get.path", "/"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.initial_delay_seconds", "60"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.period_seconds", "20"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.success_threshold", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.timeout_seconds", "10"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.liveness_probe.type", "TCP"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.basic_auth_env_var", ""),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_headers", "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_methods", "GET, PUT, POST, DELETE, PATCH, OPTIONS"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.cors_allow_origin", "*"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.denylist_source_range", ""),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.enable_cors", "true"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.enable_sticky_session", "false"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.keepalive_time_seconds", "7200"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.keepalive_timeout_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_body_size_mb", "200"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_buffer_size_kb", "3"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_connect_timeout_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_read_timeout_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.proxy_send_timeout_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.send_timeout_seconds", "120"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.network.ingress.whitelist_source_range", "127.0.0.1/0"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.failure_threshold", "12"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.http_get.path", "/"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.initial_delay_seconds", "60"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.period_seconds", "20"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.success_threshold", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.timeout_seconds", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.readiness_probe.type", "TCP"),
					resource.TestCheckResourceAttr("qovery_container.test", "advanced_settings.security.service_account_name", ""),
				),
			},
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryContainerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("container not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("container.id not found")
		}

		_, err := qoveryServices.Container.Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryContainerDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("container not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("container.id not found")
		}

		_, err := qoveryServices.Container.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found container but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted container: %s", err.Error())
		}
		return nil
	}
}

func testAccContainerDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerDefaultConfigWithName(testName string, name string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), name, containerImageName, containerTag,
	)
}

func testAccContainerDefaultConfigWithArguments(testName string, arguments []string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  arguments = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, convertStringArrayToString(arguments),
	)
}

func testAccContainerDefaultConfigWithAutoPreview(testName string, autoPreview string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  auto_preview = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, autoPreview,
	)
}

func testAccContainerDefaultConfigWithResources(testName string, cpu string, memory string, minRunningInstances string, maxRunningInstances string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  cpu = "%s"
  memory = "%s"
  min_running_instances = "%s"
  max_running_instances = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, cpu, memory, minRunningInstances, maxRunningInstances,
	)
}

func testAccContainerDefaultConfigWithStorage(testName string, storages []serviceStorage) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  storage = %s
}

`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, convertStoragesToString(storages),
	)
}
func testAccContainerDefaultConfigWithPorts(testName string, ports []servicePort) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  ports = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, convertPortsToString(ports),
	)
}

func testAccContainerDefaultConfigWithEnvironmentVariables(testName string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  environment_variables = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, convertEnvVarsToString(environmentVariables),
	)
}

func testAccContainerDefaultConfigWithAdvSettings(testName string, advSettings string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  advanced_settings = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, advSettings,
	)
}

func testAccContainerDefaultConfigWithSecrets(testName string, secrets map[string]string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
 environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  secrets = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, convertEnvVarsToString(secrets),
	)
}

//
//func testAccContainerDefaultConfigWithCustomDomains(testName string, customDomains []string, state string) string {
//	ports := []servicePort{
//		{
//			InternalPort:       8000,
//			PubliclyAccessible: true,
//			ExternalPort:       int64ToPtr(443),
//		},
//	}
//
//	return fmt.Sprintf(`
//%s
//
//resource "qovery_container" "test" {
//  environment_id = qovery_environment.test.id
//  name = "%s"
//  build_mode = "DOCKER"
//  dockerfile_path = "Dockerfile"
//  git_repository = {
//    url = "%s"
//  }
//  ports = %s
//  custom_domains = %s
//  state = "%s"
//}
//`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), containerRepositoryURL, convertPortsToString(ports), convertCustomDomainsToString(customDomains), state,
//	)
//}
//
//func testAccContainerDefaultConfigWithEnvironmentEnvVariables(testName string, environmentVariables map[string]string) string {
//	return fmt.Sprintf(`
//%s
//
//resource "qovery_container" "test" {
//  environment_id = qovery_environment.test.id
//  name = "%s"
//  build_mode = "DOCKER"
//  dockerfile_path = "Dockerfile"
//  git_repository = {
//    url = "%s"
//  }
//}
//`, testAccEnvironmentDefaultConfigWithEnvironmentVariables(testName, environmentVariables), generateTestName(testName), containerRepositoryURL,
//	)
//}
//
//func testAccContainerDefaultConfigWithDatabase(testName string) string {
//	return fmt.Sprintf(`
//%s
//
//resource "qovery_container" "test" {
//  environment_id = qovery_environment.test.id
//  name = "%s"
//  build_mode = "DOCKER"
//  dockerfile_path = "Dockerfile"
//  git_repository = {
//    url = "%s"
//  }
//}
//`, testAccDatabaseDefaultConfig(testName, redisContainer), generateTestName(testName), containerRepositoryURL,
//	)
//}

func convertStringArrayToString(array []string) string {
	return fmt.Sprintf("[\"%s\"]", strings.Join(array, "\",\""))
}

func containerAdvSettingsString() string {
	return fmt.Sprintf(`{
    "deployment.custom_domain_check_enabled" = false
    "deployment.termination_grace_period_seconds" = 120
    "hpa.cpu.average_utilization_percent" = 70
    "liveness_probe.failure_threshold" = 9
    "liveness_probe.http_get.path" ="/"
    "liveness_probe.initial_delay_seconds" = 15
    "liveness_probe.period_seconds" = 5
    "liveness_probe.success_threshold" = 3
    "liveness_probe.timeout_seconds" = 3
    "liveness_probe.type" = "TCP"
    "network.ingress.basic_auth_env_var" = ""
    "network.ingress.cors_allow_headers" = "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"
    "network.ingress.cors_allow_methods" = "GET, PUT, POST, DELETE, PATCH, OPTIONS"
    "network.ingress.cors_allow_origin" = "*"
    "network.ingress.denylist_source_range" = ""
    "network.ingress.enable_cors" = true
    "network.ingress.enable_sticky_session" = false
    "network.ingress.keepalive_time_seconds" = 1800
    "network.ingress.keepalive_timeout_seconds" = 60
    "network.ingress.proxy_body_size_mb" = 50
    "network.ingress.proxy_buffer_size_kb" = 20
    "network.ingress.proxy_connect_timeout_seconds" = 30
    "network.ingress.proxy_read_timeout_seconds" = 30
    "network.ingress.proxy_send_timeout_seconds" = 30
    "network.ingress.send_timeout_seconds" = 30
    "network.ingress.whitelist_source_range" = "0.0.0.0/0"
    "readiness_probe.failure_threshold" = 8
    "readiness_probe.http_get.path" = "/"
    "readiness_probe.initial_delay_seconds" = 15
    "readiness_probe.period_seconds" = 5
    "readiness_probe.success_threshold" = 10
    "readiness_probe.timeout_seconds" = 10
    "readiness_probe.type" = "TCP"
    "security.service_account_name" = ""
  }`)
}

func editedContainerAdvSettingsString() string {
	return fmt.Sprintf(`{
    "deployment.custom_domain_check_enabled" = true
    "deployment.termination_grace_period_seconds" = 240
    "hpa.cpu.average_utilization_percent" = 80
    "liveness_probe.failure_threshold" = 12
    "liveness_probe.http_get.path" ="/"
    "liveness_probe.initial_delay_seconds" = 60
    "liveness_probe.period_seconds" = 20
    "liveness_probe.success_threshold" = 2
    "liveness_probe.timeout_seconds" = 10
    "liveness_probe.type" = "TCP"
    "network.ingress.basic_auth_env_var" = ""
    "network.ingress.cors_allow_headers" = "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"
    "network.ingress.cors_allow_methods" = "GET, PUT, POST, DELETE, PATCH, OPTIONS"
    "network.ingress.cors_allow_origin" = "*"
    "network.ingress.denylist_source_range" = ""
    "network.ingress.enable_cors" = true
    "network.ingress.enable_sticky_session" = false
    "network.ingress.keepalive_time_seconds" = 7200
    "network.ingress.keepalive_timeout_seconds" = 120
    "network.ingress.proxy_body_size_mb" = 200
    "network.ingress.proxy_buffer_size_kb" = 3
    "network.ingress.proxy_connect_timeout_seconds" = 120
    "network.ingress.proxy_read_timeout_seconds" = 120
    "network.ingress.proxy_send_timeout_seconds" = 120
    "network.ingress.send_timeout_seconds" = 120
    "network.ingress.whitelist_source_range" = "127.0.0.1/0"
    "readiness_probe.failure_threshold" = 12
    "readiness_probe.http_get.path" = "/"
    "readiness_probe.initial_delay_seconds" = 60
    "readiness_probe.period_seconds" = 20
    "readiness_probe.success_threshold" = 2
    "readiness_probe.timeout_seconds" = 2
    "readiness_probe.type" = "TCP"
    "security.service_account_name" = ""
  }`)
}
