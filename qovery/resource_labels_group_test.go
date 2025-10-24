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

func TestAcc_LabelsGroup(t *testing.T) {
	t.Parallel()
	testName := "labels group"
	labelsName := generateTestName("labels")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryLabelsGroupDestroy(getTestOrganizationID(), "qovery_labels_group.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: getLabelsGroupConfigFromModel(
					testName,
					qovery.LabelsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(labelsName),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryLabelsGroupExists(getTestOrganizationID(), "qovery_labels_group.test"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "name", labelsName),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.key", "key1"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.value", "value1"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.propagate_to_cloud_provider", "false"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.key", "key2"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.value", "value2"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.propagate_to_cloud_provider", "true"),
				),
			},
			// Update name
			{
				Config: getLabelsGroupConfigFromModel(
					testName,
					qovery.LabelsGroup{
						OrganizationId: qovery.FromString(getTestOrganizationID()),
						Name:           qovery.FromString(labelsName + "-updated"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryLabelsGroupExists(getTestOrganizationID(), "qovery_labels_group.test"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "name", labelsName+"-updated"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.key", "key1"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.value", "value1"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.0.propagate_to_cloud_provider", "false"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.key", "key2"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.value", "value2"),
					resource.TestCheckResourceAttr("qovery_labels_group.test", "labels.1.propagate_to_cloud_provider", "true"),
				),
			},
		},
	})
}

func getLabelsGroupConfigFromModel(testName string, labelGroup qovery.LabelsGroup) string {
	tmpl_model := struct {
		LabelsGroup qovery.LabelsGroup
	}{
		LabelsGroup: labelGroup,
	}

	tmpl, err := template.New("getLabelsGroupConfigFromModel").Parse(`
data "qovery_organization" "test" {
  id = {{ .LabelsGroup.OrganizationId.String }}
}

resource "qovery_labels_group" "test" {
  organization_id = data.qovery_organization.test.id
  name = {{ .LabelsGroup.Name.String }}
  labels = [
     {
        key = "key1"
        value = "value1"
        propagate_to_cloud_provider = false
    },
    {
        key = "key2"
        value = "value2"
        propagate_to_cloud_provider = true
    }
  ]
}
`)

	var labelsGroupConfigStr bytes.Buffer
	err = tmpl.Execute(&labelsGroupConfigStr, tmpl_model)
	if err != nil {
		return ""
	}

	return labelsGroupConfigStr.String()
}

func testAccQoveryLabelsGroupDestroy(organizationId string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("labels group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("labels_group.id not found")
		}

		_, err := qoveryServices.LabelsGroup.Get(context.TODO(), organizationId, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found labels goup but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted labels group: %s", err.Error())
		}
		return nil
	}
}

func testAccQoveryLabelsGroupExists(organizationId string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("labels group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("labels.id not found")
		}

		_, err := qoveryServices.LabelsGroup.Get(context.TODO(), organizationId, rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}
