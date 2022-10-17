//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

const (
	applicationRepositoryURL = "https://github.com/Qovery/test_http_server.git"
	applicationBranch        = "master"
)

type serviceStorage struct {
	Type       string
	Size       int64
	MountPoint string
}

func (s serviceStorage) String() string {
	return fmt.Sprintf(`
{
  type = "%s"
  size = %d
  mount_point = "%s"
}
`, s.Type, s.Size, s.MountPoint)
}

type servicePort struct {
	InternalPort       int64
	PubliclyAccessible bool
	Name               *string
	ExternalPort       *int64
	Protocol           *string
}

func (p servicePort) String() string {

	str := fmt.Sprintf(`
{
  internal_port = %d
  publicly_accessible = "%t"
`, p.InternalPort, p.PubliclyAccessible)
	if p.Name != nil {
		str += fmt.Sprintf("  name = \"%s\"\n", *p.Name)
	}
	if p.ExternalPort != nil {
		str += fmt.Sprintf("  external_port = %d\n", *p.ExternalPort)
	}
	if p.Protocol != nil {
		str += fmt.Sprintf("  protocol = \"%s\"\n", *p.Protocol)
	}
	str += "}"
	return str
}

func TestAcc_Application(t *testing.T) {
	t.Parallel()
	testName := "application"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", "master"),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update name
			{
				Config: testAccApplicationDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update auto_preview
			{
				Config: testAccApplicationDefaultConfigWithAutoPreview(
					testName,
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update resources
			{
				Config: testAccApplicationDefaultConfigWithResources(
					testName,
					"1000",
					"1024",
					"2",
					"3",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "1000"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "1024"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "2"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "3"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ApplicationWithAutoPreview(t *testing.T) {
	t.Parallel()
	testName := "application-with-auto-preview"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithAutoPreview(
					testName,
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update auto_preview
			{
				Config: testAccApplicationDefaultConfigWithAutoPreview(
					testName,
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithAutoPreview(
					testName,
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ApplicationWithState(t *testing.T) {
	t.Parallel()
	testName := "application-with-state"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithState(
					testName,
					"STOPPED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "STOPPED"),
				),
			},
			// Update state
			{
				Config: testAccApplicationDefaultConfigWithState(
					testName,
					"RUNNING",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ApplicationWithEnvironmentVariables(t *testing.T) {
	t.Parallel()
	testName := "application-with-environment-variables"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update environment variable
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Add environment variable
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Remove environment variables
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentVariables(
					testName,
					map[string]string{
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ApplicationWithSecrets(t *testing.T) {
	t.Parallel()
	testName := "application-with-secrets"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update secret
			{
				Config: testAccApplicationDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Add secret
			{
				Config: testAccApplicationDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Remove secret
			{
				Config: testAccApplicationDefaultConfigWithSecrets(
					testName,
					map[string]string{
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
		},
	})
}

func TestAcc_ApplicationWithCustomDomains(t *testing.T) {
	t.Parallel()
	testName := "application-with-custom-domains"
	customDomain := gofakeit.DomainName()
	updatedCustomDomain := gofakeit.DomainName()

	// NOTE: Run this test with stopped application unless we are in the main branch.
	// Running it with a STOPPED state will make the test run way faster.
	state := "STOPPED"
	if isCIMainBranch() {
		state = "RUNNING"
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfigWithCustomDomains(
					testName,
					[]string{customDomain},
					state,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "8000"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "custom_domains.*", map[string]string{
						"domain": customDomain,
					}),
					resource.TestCheckResourceAttrSet("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", state),
				),
			},
			// Update environment variable
			{
				Config: testAccApplicationDefaultConfigWithCustomDomains(
					testName,
					[]string{updatedCustomDomain},
					state,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "8000"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "custom_domains.*", map[string]string{
						"domain": updatedCustomDomain,
					}),
					resource.TestCheckResourceAttrSet("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", state),
				),
			},
			// Add environment variable
			{
				Config: testAccApplicationDefaultConfigWithCustomDomains(
					testName,
					[]string{customDomain, updatedCustomDomain},
					state,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "8000"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "custom_domains.*", map[string]string{
						"domain": customDomain,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "custom_domains.*", map[string]string{
						"domain": updatedCustomDomain,
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckResourceAttrSet("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", state),
				),
			},
			// Remove environment variables
			{
				Config: testAccApplicationDefaultConfigWithCustomDomains(
					testName,
					[]string{customDomain},
					state,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "8000"),
					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "true"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "custom_domains.*", map[string]string{
						"domain": customDomain,
					}),
					resource.TestCheckResourceAttrSet("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", state),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Application should restart when environment env variables are updated.
func TestAcc_ApplicationRestartOnEnvironmentUpdate(t *testing.T) {
	t.Parallel()
	testName := "application-restart"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update environment env variables
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentEnvVariables(
					testName,
					map[string]string{
						"key1": "value1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Update environment variables
			{
				Config: testAccApplicationDefaultConfigWithEnvironmentEnvVariables(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TODO: uncomment after debugging why storage can't be updated
//func TestAcc_ApplicationWithStorage(t *testing.T) {
//	t.Parallel()
//	testName := "application-with-storage"
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccApplicationDefaultConfigWithStorage(
//					testName,
//					[]serviceStorage{
//						{
//							Type:       "FAST_SSD",
//							Size:       1,
//							MountPoint: "/data",
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.type", "FAST_SSD"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.size", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.mount_point", "/"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//			// Add another storage
//			{
//				Config: testAccApplicationDefaultConfigWithStorage(
//					testName,
//					[]serviceStorage{
//						{
//							Type:       "FAST_SSD",
//							Size:       1,
//							MountPoint: "/toto",
//						},
//						{
//							Type:       "FAST_SSD",
//							Size:       1,
//							MountPoint: "/tata",
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.type", "FAST_SSD"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.size", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.mount_point", "/toto"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.1.type", "FAST_SSD"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.1.size", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.1.mount_point", "/tata"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//			// Remove first storage
//			{
//				Config: testAccApplicationDefaultConfigWithStorage(
//					testName,
//					[]serviceStorage{
//						{
//							Type:       "FAST_SSD",
//							Size:       1,
//							MountPoint: "/toto",
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.type", "FAST_SSD"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.size", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "storage.0.mount_point", "/toto"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//		},
//	})
//}

// TODO: uncomment after debugging why ports can't be updated
//func TestAcc_ApplicationWithPorts(t *testing.T) {
//	t.Parallel()
//	testName := "application-with-ports"
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccApplicationDefaultConfigWithPorts(
//					testName,
//					[]servicePort{
//						{
//							InternalPort:       80,
//							PubliclyAccessible: false,
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "80"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "false"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//			// Add another port
//			{
//				Config: testAccApplicationDefaultConfigWithPorts(
//					testName,
//					[]servicePort{
//						{
//							InternalPort:       80,
//							PubliclyAccessible: false,
//						},
//						{
//							Name:               stringToPtr("external port"),
//							InternalPort:       81,
//							ExternalPort:       int64ToPtr(443),
//							PubliclyAccessible: true,
//							Protocol:           stringToPtr("HTTP"),
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "80"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "false"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.1.name", "external port"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.1.internal_port", "81"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.1.external_port", "443"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.1.publicly_accessible", "true"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.1.protocol", "HTTP"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//			// Remove first port
//			{
//				Config: testAccApplicationDefaultConfigWithPorts(
//					testName,
//					[]servicePort{
//						{
//							Name:               stringToPtr("external port"),
//							InternalPort:       81,
//							ExternalPort:       int64ToPtr(443),
//							PubliclyAccessible: true,
//							Protocol:           stringToPtr("HTTP"),
//						},
//					},
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryProjectExists("qovery_project.test"),
//					testAccQoveryEnvironmentExists("qovery_environment.test"),
//					testAccQoveryApplicationExists("qovery_application.test"),
//					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
//					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
//					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
//					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
//					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
//					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
//					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
//					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.name", "external port"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.internal_port", "81"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.external_port", "443"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.publicly_accessible", "true"),
//					resource.TestCheckResourceAttr("qovery_application.test", "ports.0.protocol", "HTTP"),
//					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
//					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
//						"key": regexp.MustCompile(`^QOVERY_`),
//					}),
//					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
//				),
//			},
//		},
//	})
//}

func TestAcc_ApplicationImport(t *testing.T) {
	t.Parallel()
	testName := "application-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.branch", applicationBranch),
					resource.TestCheckResourceAttr("qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "ports.0"),
					resource.TestCheckNoResourceAttr("qovery_application.test", "environment_variables.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_application.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_application.test", "internal_host", regexp.MustCompile(`^app-z`)),
					resource.TestCheckResourceAttr("qovery_application.test", "state", "RUNNING"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryApplicationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("application not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("application.id not found")
		}

		_, apiErr := apiClient.GetApplication(context.TODO(), rs.Primary.ID)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryApplicationDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("application not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("application.id not found")
		}

		_, apiErr := apiClient.GetApplication(context.TODO(), rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("found application but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted application: %s", apiErr.Summary())
		}
		return nil
	}
}

func testAccApplicationDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL,
	)
}

func testAccApplicationDefaultConfigWithState(testName string, state string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  state = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, state,
	)
}

func testAccApplicationDefaultConfigWithAutoPreview(testName string, autoPreview string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  auto_preview = "%s"
  git_repository = {
    url = "%s"
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), autoPreview, applicationRepositoryURL,
	)
}

func testAccApplicationDefaultConfigWithResources(testName string, cpu string, memory string, minRunningInstances string, maxRunningInstances string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  cpu = "%s"
  memory = "%s"
  min_running_instances = "%s"
  max_running_instances = "%s"
  git_repository = {
    url = "%s"
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), cpu, memory, minRunningInstances, maxRunningInstances, applicationRepositoryURL,
	)
}

func testAccApplicationDefaultConfigWithStorage(testName string, storages []serviceStorage) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  storage = %s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, convertStoragesToString(storages),
	)
}

func testAccApplicationDefaultConfigWithPorts(testName string, ports []servicePort) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  ports = %s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, convertPortsToString(ports),
	)
}

func testAccApplicationDefaultConfigWithEnvironmentVariables(testName string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  environment_variables = %s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, convertEnvVarsToString(environmentVariables),
	)
}

func testAccApplicationDefaultConfigWithSecrets(testName string, secrets map[string]string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  secrets = %s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, convertEnvVarsToString(secrets),
	)
}

func testAccApplicationDefaultConfigWithCustomDomains(testName string, customDomains []string, state string) string {
	ports := []servicePort{
		{
			InternalPort:       8000,
			PubliclyAccessible: true,
			ExternalPort:       int64ToPtr(443),
		},
	}

	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
  ports = %s
  custom_domains = %s
  state = "%s"
} 
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, convertPortsToString(ports), convertCustomDomainsToString(customDomains), state,
	)
}

func testAccApplicationDefaultConfigWithEnvironmentEnvVariables(testName string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
}
`, testAccEnvironmentDefaultConfigWithEnvironmentVariables(testName, environmentVariables), generateTestName(testName), applicationRepositoryURL,
	)
}

func testAccApplicationDefaultConfigWithDatabase(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
  }
}
`, testAccDatabaseDefaultConfig(testName, redisContainer), generateTestName(testName), applicationRepositoryURL,
	)
}

func convertStoragesToString(storages []serviceStorage) string {
	storagesStr := make([]string, 0, len(storages))
	for _, storage := range storages {
		storagesStr = append(storagesStr, storage.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(storagesStr, ","))
}

func convertPortsToString(ports []servicePort) string {
	portsStr := make([]string, 0, len(ports))
	for _, port := range ports {
		portsStr = append(portsStr, port.String())
	}
	ret := fmt.Sprintf("[%s]", strings.Join(portsStr, ","))
	fmt.Println(ret)
	return fmt.Sprintf("[%s]", strings.Join(portsStr, ","))
}

func convertCustomDomainsToString(customDomains []string) string {
	domains := make([]string, 0, len(customDomains))
	for _, domain := range customDomains {
		domains = append(domains, fmt.Sprintf(`{domain: "%s"}`, domain))
	}
	return fmt.Sprintf("[%s]", strings.Join(domains, ","))
}

func stringToPtr(v string) *string {
	return &v
}

func int64ToPtr(v int64) *int64 {
	return &v
}
