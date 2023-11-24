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

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

func TestAcc_DatabaseContainer(t *testing.T) {
	t.Parallel()
	testName := "database-container"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("6.2"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6.2"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"), // not set because container
				),
			},
			// Update name
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName) + "-updated"),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("6.2"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", fmt.Sprintf("%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6.2"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"), // not set because container
				),
			},
			// Update accessibility
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("6.2"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PRIVATE"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6.2"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PRIVATE"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"), // not set because container
				),
			},
			// Update resources
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("6.2"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(500),
						Memory:        qovery.FromInt32(512),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6.2"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"), // not set because container
				),
			},
			// Update version
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("7.0"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(500),
						Memory:        qovery.FromInt32(512),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "7.0"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"), // not set because container
				),
			},
		},
	})
}

func TestAcc_DatabaseManaged(t *testing.T) {
	skipInCIUnlessMainBranch(t)
	t.Parallel()
	testName := "database-managed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
			// Update name
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName) + "-updated"),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", fmt.Sprintf("%s-updated", generateTestName(testName))),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
			// Update accessibility
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PRIVATE"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PRIVATE"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
			// Update resources
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(500),
						Memory:        qovery.FromInt32(512),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
			// Update storage
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(15),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "15"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
			// Update instance type
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("13"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(15),
						InstanceType:  qovery.FromString("db.t3.small"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "15"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.small"),
				),
			},
			// Update version
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("POSTGRESQL"),
						Version:       qovery.FromString("14"),
						Mode:          qovery.FromString("MANAGED"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(15),
						InstanceType:  qovery.FromString("db.t3.micro"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "14"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "15"),
					resource.TestCheckResourceAttr("qovery_database.test", "instance_type", "db.t3.micro"),
				),
			},
		},
	})
}

func TestAcc_DatabaseImport(t *testing.T) {
	t.Parallel()
	testName := "database-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: GetDatabaseConfigFromModel(
					testName,
					qovery.Database{
						Name:          qovery.FromString(generateTestName(testName)),
						Type:          qovery.FromString("REDIS"),
						Version:       qovery.FromString("6.2"),
						Mode:          qovery.FromString("CONTAINER"),
						Accessibility: qovery.FromString("PUBLIC"),
						CPU:           qovery.FromInt32(250),
						Memory:        qovery.FromInt32(256),
						Storage:       qovery.FromInt32(10),
						InstanceType:  qovery.FromStringPointer(nil),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6.2"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckNoResourceAttr("qovery_database.test", "instance_type"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryDatabaseExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("database not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("database.id not found")
		}

		_, apiErr := apiClient.GetDatabase(context.TODO(), rs.Primary.ID)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryDatabaseDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("database not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("database.id not found")
		}

		_, apiErr := apiClient.GetDatabase(context.TODO(), rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("found database but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted database: %s", apiErr.Summary())
		}
		return nil
	}
}

func GetDatabaseConfigFromModel(testName string, db qovery.Database) string {
	tmpl_model := struct {
		EnvironmentStr string
		Database       qovery.Database
	}{
		EnvironmentStr: testAccEnvironmentDefaultConfig(testName),
		Database:       db,
	}

	tmpl, err := template.New("GetDatabaseConfigFromModel").Parse(`
{{ .EnvironmentStr }}

resource "qovery_database" "test" {
	environment_id = qovery_environment.test.id
	name = {{ .Database.Name.String }}
	type = {{ .Database.Type.String }}
	version = {{ .Database.Version.String }}
	mode = {{ .Database.Mode.String }}

	{{ with .Database.InstanceType }}
	{{ if not .IsNull }}
	instance_type = {{ .String }}
	{{ end }}
	{{ end }}
	{{ with .Database.Accessibility }}
	accessibility = {{ .String }}
	{{ end }}
	{{ with .Database.CPU }}
	cpu = {{ .ValueInt64 }}
	{{ end }}
	{{ with .Database.Memory }}
	memory = {{ .ValueInt64 }}
	{{ end }}
	{{ with .Database.Storage }}
	storage = {{ .ValueInt64 }}
	{{ end }}
}
`)

	var jobConfigStr bytes.Buffer
	err = tmpl.Execute(&jobConfigStr, tmpl_model)
	if err != nil {
		return ""
	}

	return jobConfigStr.String()
}
