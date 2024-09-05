//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

const ()

func generateHelmVariableSet(key string, value string) types.Set {
	var values = types.ObjectValueMust(attributes, map[string]attr.Value{
		"key":   qovery.FromString(key),
		"value": qovery.FromString(value),
	})
	var elements = make([]attr.Value, 0, 1)
	elements = append(elements, values)
	return types.SetValueMust(
		variableObjectType,
		elements,
	)
}

func TestAcc_Helm(t *testing.T) {
	t.Parallel()
	testName := "helm"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryHelmDestroy("qovery_helm.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName)),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName))),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride: &qovery.HelmValuesOverride{
							HelmValuesOverrideFile: nil,
						},
						EnvironmentVariables:         generateHelmVariableSet("key1", "value1"),
						EnvironmentVariableAliases:   generateHelmVariableSet("key1_alias", "key1"),
						EnvironmentVariableOverrides: generateHelmVariableSet("environment_variable", "override value"),
						Secrets:                      generateHelmVariableSet("secretkey1", "secretvalue1"),
						AdvancedSettingsJson:         qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryHelmExists("qovery_helm.test"),
					resource.TestCheckResourceAttr("qovery_helm.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_helm.test", "icon_uri", fmt.Sprintf("app://qovery-console/%s", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "timeout_sec", "600"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_deploy", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "allow_cluster_wide_resources", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_name", "httpbin"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_version", "1.0.0"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "values_override.file"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "secretvalue1",
					}),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_helm.test", "internal_host", regexp.MustCompile(`^helm-z`)),
					resource.TestCheckResourceAttr("qovery_helm.test", "advanced_settings_json", "{\"network.ingress.proxy_buffer_size_kb\":8}"),
				),
			},
			// Update name
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName) + "-updated"),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName))),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride:       &qovery.HelmValuesOverride{},
						EnvironmentVariables: generateHelmVariableSet("key1", ""),
						Secrets:              generateHelmVariableSet("secretkey1", ""),
						AdvancedSettingsJson: qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryHelmExists("qovery_helm.test"),
					resource.TestCheckResourceAttr("qovery_helm.test", "name", fmt.Sprintf("%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "icon_uri", fmt.Sprintf("app://qovery-console/%s", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "timeout_sec", "600"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_deploy", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "allow_cluster_wide_resources", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_name", "httpbin"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_version", "1.0.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "",
					}),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_helm.test", "internal_host", regexp.MustCompile(`^helm-z`)),
					resource.TestCheckResourceAttr("qovery_helm.test", "advanced_settings_json", "{\"network.ingress.proxy_buffer_size_kb\":8}"),
				),
			},
			// Update iconUri
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName) + "-updated"),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName)+"-updated")),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride:       &qovery.HelmValuesOverride{},
						EnvironmentVariables: generateHelmVariableSet("key1", ""),
						Secrets:              generateHelmVariableSet("secretkey1", ""),
						AdvancedSettingsJson: qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryHelmExists("qovery_helm.test"),
					resource.TestCheckResourceAttr("qovery_helm.test", "name", fmt.Sprintf("%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "icon_uri", fmt.Sprintf("app://qovery-console/%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "timeout_sec", "600"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_deploy", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "allow_cluster_wide_resources", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_name", "httpbin"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_version", "1.0.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "",
					}),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_helm.test", "internal_host", regexp.MustCompile(`^helm-z`)),
					resource.TestCheckResourceAttr("qovery_helm.test", "advanced_settings_json", "{\"network.ingress.proxy_buffer_size_kb\":8}"),
				),
			},
			// Update variables
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName) + "-updated"),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName)+"-updated")),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride:       &qovery.HelmValuesOverride{},
						EnvironmentVariables: generateHelmVariableSet("key1", "updated"),
						Secrets:              generateHelmVariableSet("secretkey1", "updated"),
						AdvancedSettingsJson: qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "updated",
					}),
				),
			},
			// Delete variables
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName) + "-updated"),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName)+"-updated")),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride:       &qovery.HelmValuesOverride{},
						AdvancedSettingsJson: qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "environment_variables.*"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "environment_variable_aliases.0"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "environment_variable_overrides.0"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "secrets.*"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "secret_aliases.0"),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "secret_overrides.0"),
				),
			},
			// Create and Read testing with raw file as values override
			{
				Config: getHelmConfigFromModel(
					testName,
					qovery.Helm{
						Name:                      qovery.FromString(generateTestName(testName) + "-values-file-raw"),
						IconUri:                   qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName)+"-updated")),
						TimeoutSec:                qovery.FromInt64(600),
						AutoPreview:               qovery.FromBool(false),
						AutoDeploy:                qovery.FromBool(false),
						AllowClusterWideResources: qovery.FromBool(false),
						Source: &qovery.HelmSource{
							HelmSourceHelmRepository: &qovery.HelmSourceHelmRepository{
								HelmChartName:    qovery.FromString("httpbin"),
								HelmChartVersion: qovery.FromString("1.0.0"),
							},
							HelmSourceGitRepository: nil,
						},
						ValuesOverride: &qovery.HelmValuesOverride{
							HelmValuesOverrideFile: &qovery.HelmValuesOverrideFile{
								Raw: &map[string]qovery.HelmValuesRaw{},
							},
						},
						EnvironmentVariables:         generateHelmVariableSet("key1", "value1"),
						EnvironmentVariableAliases:   generateHelmVariableSet("key1_alias", "key1"),
						EnvironmentVariableOverrides: generateHelmVariableSet("environment_variable", "override value"),
						Secrets:                      generateHelmVariableSet("secretkey1", "secretvalue1"),
						AdvancedSettingsJson:         qovery.FromString("{\"network.ingress.proxy_buffer_size_kb\":8}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryHelmExists("qovery_helm.test"),
					resource.TestCheckResourceAttr("qovery_helm.test", "name", generateTestName(testName)+"-values-file-raw"),
					resource.TestCheckResourceAttr("qovery_helm.test", "icon_uri", fmt.Sprintf("app://qovery-console/%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_helm.test", "timeout_sec", "600"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "auto_deploy", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "allow_cluster_wide_resources", "false"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_name", "httpbin"),
					resource.TestCheckResourceAttr("qovery_helm.test", "source.helm_repository.chart_version", "1.0.0"),
					resource.TestCheckResourceAttr("qovery_helm.test", "values_override.file.raw.file1.content", "content"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_helm.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_helm.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "secretvalue1",
					}),
					resource.TestCheckNoResourceAttr("qovery_helm.test", "external_host"),
					resource.TestMatchResourceAttr("qovery_helm.test", "internal_host", regexp.MustCompile(`^helm-z`)),
					resource.TestCheckResourceAttr("qovery_helm.test", "advanced_settings_json", "{\"network.ingress.proxy_buffer_size_kb\":8}"),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_helm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secrets", "secret_aliases", "secret_overrides"},
			},
		},
	})
}

func getHelmConfigFromModel(testName string, helm qovery.Helm) string {
	tmpl_model := struct {
		EnvironmentStr    string
		HelmRepositoryStr string
		Helm              qovery.Helm
	}{
		EnvironmentStr:    testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndSecrets(testName, map[string]string{"environment_variable": "simple value"}, map[string]string{"environment_secret": "simple value"}),
		HelmRepositoryStr: testAccHelmRepositoryConfig(testName, "https://gitlab.com/mulesoft-int/helm-repository/-/raw/master/", "HTTPS"),
		Helm:              helm,
	}

	tmpl, err := template.New("getHelmConfigFromModel").Parse(`
{{ .EnvironmentStr }}

{{ .HelmRepositoryStr }}

resource "qovery_helm" "test" {
  environment_id = qovery_environment.test.id
  name = {{ .Helm.Name.String }}
  icon_uri = {{ .Helm.IconUri.String }}
  timeout_sec = {{ .Helm.TimeoutSec }}
  auto_preview = {{ .Helm.AutoPreview }}
  auto_deploy = {{ .Helm.AutoDeploy }}
  allow_cluster_wide_resources = {{ .Helm.AllowClusterWideResources }}

  {{ if not .Helm.EnvironmentVariables.IsNull }}
  environment_variables = {{ .Helm.EnvironmentVariables.String }}
  {{ end }}

  {{ if not .Helm.EnvironmentVariableAliases.IsNull }}
  environment_variable_aliases = {{ .Helm.EnvironmentVariableAliases.String }}
  {{ end }}

  {{ if not .Helm.EnvironmentVariableOverrides.IsNull }}
  environment_variable_overrides = {{ .Helm.EnvironmentVariableOverrides.String }}
  {{ end }}

  {{ if not .Helm.Secrets.IsNull }}	
  secrets = {{ .Helm.Secrets.String }}	
  {{ end }}

  {{ if not .Helm.SecretAliases.IsNull }}
  secret_aliases = {{ .Helm.SecretAliases.String }}
  {{ end }}

  {{ if not .Helm.SecretOverrides.IsNull }}
  secret_overrides = {{ .Helm.SecretOverrides.String }}
  {{ end }}


  {{ if not .Helm.AdvancedSettingsJson.IsNull }}
  advanced_settings_json = jsonencode({{ .Helm.AdvancedSettingsJson.ValueString }})
  {{ end }}

  {{ with .Helm.Source }}	
  source = {
	{{ with .HelmSourceHelmRepository }}	
    helm_repository = {
      helm_repository_id = qovery_helm_repository.test.id
      chart_name = {{ .HelmChartName.String }}
      chart_version = {{ .HelmChartVersion.String }}
    }
    {{ end }}
	{{ with .HelmSourceGitRepository }}	
	git_repository = {
		{{ with .Url }}
		url = {{ .String }}
		{{ end }}
		{{ with .Branch }}
		branch = {{ .String }}
		{{ end }}
		{{ with .RootPath }}
		root_path = {{ .String }}
		{{ end }}
		{{ with .GitTokenId }}
		git_token_id = {{ .String }}
		{{ end }}
	}
	{{ end }}
  }
  {{ end }}

  {{ with .Helm.ValuesOverride }}	
  values_override = {
	"set"= {}
	"set_string"= {}
	"set_json"= {}
  	{{ with .HelmValuesOverrideFile }}	
   	file= {
		{{ with .Raw }}	
		raw = { 
			file1 = { 
				content = "content"
			}
		}
	 	{{ end }}
	 	{{ with .GitRepository }}	
	 	git_repository = { 
			url = "https://github.com/Qovery/helm_chart_engine_testing.git"
            branch = "main"
		   	paths = [ "/simple_app/values.yaml" ]
		}
        {{ end }}
  	}   
  	{{ end }}
  }
  {{ end }}
}
`)

	var helmConfigStr bytes.Buffer
	err = tmpl.Execute(&helmConfigStr, tmpl_model)
	if err != nil {
		return ""
	}

	return helmConfigStr.String()
}

func testAccQoveryHelmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("helm not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("helm.id not found")
		}

		_, err := qoveryServices.Helm.Get(context.TODO(), rs.Primary.ID, "{}", false)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryHelmDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("helm not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("helm.id not found")
		}

		_, err := qoveryServices.Helm.Get(context.TODO(), rs.Primary.ID, "{}", false)
		if err == nil {
			return fmt.Errorf("found helm but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted helm: %s", err.Error())
		}
		return nil
	}
}
