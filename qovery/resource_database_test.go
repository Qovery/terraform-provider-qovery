package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_DatabaseContainer(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseDefaultConfig(
					generateDatabaseName(nameSuffix),
					"REDIS",
					"6",
					"CONTAINER",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update name
			{
				Config: testAccDatabaseDefaultConfig(
					fmt.Sprintf("%s-updated", generateDatabaseName(nameSuffix)),
					"REDIS",
					"6",
					"CONTAINER",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", fmt.Sprintf("%s-updated", generateDatabaseName(nameSuffix))),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update accessibility
			{
				Config: testAccDatabaseDefaultConfigWithAccessibility(
					generateDatabaseName(nameSuffix),
					"REDIS",
					"6",
					"CONTAINER",
					"PRIVATE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PRIVATE"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update resources
			{
				Config: testAccDatabaseDefaultConfigWithResources(
					generateDatabaseName(nameSuffix),
					"REDIS",
					"6",
					"CONTAINER",
					500,
					512,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// TODO: uncomment after debugging why storage can't be updated
			//// Update storage
			//{
			//	Config: testAccDatabaseDefaultConfigWithStorage(
			//		generateDatabaseName(nameSuffix),
			//		"REDIS",
			//		"6",
			//		"CONTAINER",
			//		15,
			//	),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		testAccQoveryDatabaseExists("qovery_database.test"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
			//		resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
			//		resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "storage", "15"),
			//		resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
			//	),
			//},
		},
	})
}

func TestAcc_DatabaseManaged(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseDefaultConfig(
					generateDatabaseName(nameSuffix),
					"POSTGRESQL",
					"13",
					"MANAGED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update name
			{
				Config: testAccDatabaseDefaultConfig(
					fmt.Sprintf("%s-updated", generateDatabaseName(nameSuffix)),
					"POSTGRESQL",
					"13",
					"MANAGED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", fmt.Sprintf("%s-updated", generateDatabaseName(nameSuffix))),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update accessibility
			{
				Config: testAccDatabaseDefaultConfigWithAccessibility(
					generateDatabaseName(nameSuffix),
					"POSTGRESQL",
					"13",
					"MANAGED",
					"PRIVATE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PRIVATE"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update resources
			{
				Config: testAccDatabaseDefaultConfigWithResources(
					generateDatabaseName(nameSuffix),
					"POSTGRESQL",
					"13",
					"MANAGED",
					500,
					512,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update storage
			{
				Config: testAccDatabaseDefaultConfigWithStorage(
					generateDatabaseName(nameSuffix),
					"POSTGRESQL",
					"13",
					"MANAGED",
					15,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "POSTGRESQL"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "13"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "15"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
		},
	})
}

func TestAcc_DatabaseImport(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseDefaultConfig(
					generateDatabaseName(nameSuffix),
					"REDIS",
					"6",
					"CONTAINER",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateDatabaseName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
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

func testAccDatabaseDefaultConfig(name string, dbType string, version string, mode string) string {
	return fmt.Sprintf(`
resource "qovery_database" "test" {
  environment_id = "%s"
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
}
`, getTestEnvironmentID(), name, dbType, version, mode,
	)
}

func testAccDatabaseDefaultConfigWithAccessibility(name string, dbType string, version string, mode string, accessibility string) string {
	return fmt.Sprintf(`
resource "qovery_database" "test" {
  environment_id = "%s"
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  accessibility = "%s"
}
`, getTestEnvironmentID(), name, dbType, version, mode, accessibility,
	)
}

func testAccDatabaseDefaultConfigWithResources(name string, dbType string, version string, mode string, cpu int64, memory int64) string {
	return fmt.Sprintf(`
resource "qovery_database" "test" {
  environment_id = "%s"
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  cpu = %d
  memory = %d
}
`, getTestEnvironmentID(), name, dbType, version, mode, cpu, memory,
	)
}

func testAccDatabaseDefaultConfigWithStorage(name string, dbType string, version string, mode string, storage int64) string {
	return fmt.Sprintf(`
resource "qovery_database" "test" {
  environment_id = "%s"
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  storage = %d
}
`, getTestEnvironmentID(), name, dbType, version, mode, storage,
	)
}

func generateDatabaseName(suffix string) string {
	return fmt.Sprintf("%s-database-%s", testResourcePrefix, suffix)
}
