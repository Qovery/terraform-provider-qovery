//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

const (
	test_annotations_name = "annotations for testing terraform provider"
)

func generateAnnotationsMap(keysAndValues ...string) map[string]string {
	if len(keysAndValues)%2 != 0 {
		panic("keysAndValues error")
	}

	annotations := make(map[string]string)
	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		value := keysAndValues[i+1]
		annotations[key] = value
	}

	return annotations
}

func generateScopesList(values ...string) []string {
	return append([]string{}, values...)
}

func TestAcc_AnnotationsGroup(t *testing.T) {
	t.Parallel()
	testName := "annotations group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryAnnotationsGroupDestroy(getTestOrganizationID(), "qovery_annotations_group.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getAnnotationsGroupConfigFromModel(
					testName,
					qovery.AnnotationsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(test_annotations_name),
						Annotations:    generateAnnotationsMap("key1", "value1"),
						Scopes:         append([]string{}, "PODS"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAnnotationsGroupExists(getTestOrganizationID(), "qovery_annotations_group.test"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "name", test_annotations_name),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "annotations.key1", "value1"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "scopes.0", "PODS"),
				),
			},
			// Update name
			{
				Config: getAnnotationsGroupConfigFromModel(
					testName,
					qovery.AnnotationsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(test_annotations_name + "-updated"),
						Annotations:    generateAnnotationsMap("key1", "value1"),
						Scopes:         append([]string{}, "PODS"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAnnotationsGroupExists(getTestOrganizationID(), "qovery_annotations_group.test"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "name", test_annotations_name+"-updated"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "annotations.key1", "value1"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "scopes.0", "PODS"),
				),
			},
			//Update Scopes
			{
				Config: getAnnotationsGroupConfigFromModel(
					testName,
					qovery.AnnotationsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(test_annotations_name + "-updated"),
						Annotations:    generateAnnotationsMap("key1", "value1"),
						Scopes:         append([]string{}, "DEPLOYMENTS"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAnnotationsGroupExists(getTestOrganizationID(), "qovery_annotations_group.test"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "name", test_annotations_name+"-updated"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "annotations.key1", "value1"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "scopes.0", "DEPLOYMENTS"),
				),
			},
			//Update Annotations
			{
				Config: getAnnotationsGroupConfigFromModel(
					testName,
					qovery.AnnotationsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(test_annotations_name + "-updated"),
						Annotations:    generateAnnotationsMap("key1", "value1", "key2", "value2"),
						Scopes:         append([]string{}, "DEPLOYMENTS"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAnnotationsGroupExists(getTestOrganizationID(), "qovery_annotations_group.test"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "name", test_annotations_name+"-updated"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "annotations.key1", "value1"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "annotations.key2", "value2"),
					resource.TestCheckResourceAttr("qovery_annotations_group.test", "scopes.0", "DEPLOYMENTS"),
				),
			},
		},
	})
}

func getAnnotationsGroupConfigFromModel(testName string, annotationGroup qovery.AnnotationsGroup) string {
	tmpl_model := struct {
		AnnotationsGroup qovery.AnnotationsGroup
	}{
		AnnotationsGroup: annotationGroup,
	}

	tmpl, err := template.New("getAnnotationsGroupConfigFromModel").Parse(`
data "qovery_organization" "test" {
  id = {{ .AnnotationsGroup.OrganizationId.String }}
}

resource "qovery_annotations_group" "test" {
  organization_id = data.qovery_organization.test.id
  name = {{ .AnnotationsGroup.Name.String }}
  annotations = {
    {{- range $key, $value := .AnnotationsGroup.Annotations }}
    {{ $key }} = "{{ $value }}"
    {{- end }}
  }
  scopes = [
    {{- range $i, $scope := .AnnotationsGroup.Scopes }}
    {{ if $i }}, {{ end }}"{{ $scope }}"
    {{- end }}
  ]
}
`)

	var annotationsGroupConfigStr bytes.Buffer
	err = tmpl.Execute(&annotationsGroupConfigStr, tmpl_model)
	if err != nil {
		return ""
	}

	return annotationsGroupConfigStr.String()
}

func testAccQoveryAnnotationsGroupDestroy(organizationId string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("annotations group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("annotations_group.id not found")
		}

		_, err := qoveryServices.AnnotationsGroup.Get(context.TODO(), organizationId, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found annotations goup but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted annotations group: %s", err.Error())
		}
		return nil
	}
}

func testAccQoveryAnnotationsGroupExists(organizationId string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("annotations group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("annotations.id not found")
		}

		_, err := qoveryServices.AnnotationsGroup.Get(context.TODO(), organizationId, rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}
