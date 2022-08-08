//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type database struct {
	dbType  string
	version string
	mode    string
}

var redisContainer = database{
	dbType:  "REDIS",
	version: "6",
	mode:    "CONTAINER",
}

var psqlManaged = database{
	dbType:  "POSTGRESQL",
	version: "13",
	mode:    "MANAGED",
}

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
				Config: testAccDatabaseDefaultConfig(
					testName,
					redisContainer,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
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
					fmt.Sprintf("%s-updated", testName),
					redisContainer,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
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
					testName,
					redisContainer,
					"PRIVATE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
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
					testName,
					redisContainer,
					500,
					512,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
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
				Config: testAccDatabaseDefaultConfig(
					testName,
					psqlManaged,
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
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update name
			{
				Config: testAccDatabaseDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					psqlManaged,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
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
					testName,
					psqlManaged,
					"PRIVATE",
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
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update resources
			{
				Config: testAccDatabaseDefaultConfigWithResources(
					testName,
					psqlManaged,
					500,
					512,
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
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
			// Update storage
			{
				Config: testAccDatabaseDefaultConfigWithStorage(
					testName,
					psqlManaged,
					15,
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
					resource.TestCheckResourceAttr("qovery_database.test", "state", "RUNNING"),
				),
			},
		},
	})
}

func TestAcc_DatabaseWithState(t *testing.T) {
	t.Parallel()
	testName := "database-with-state"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseDefaultConfigWithState(
					testName,
					redisContainer,
					"STOPPED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("qovery_database.test", "accessibility", "PUBLIC"),
					resource.TestCheckResourceAttr("qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("qovery_database.test", "state", "STOPPED"),
				),
			},
			// Update state
			{
				Config: testAccDatabaseDefaultConfigWithState(
					testName,
					redisContainer,
					"RUNNING",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
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
				Config: testAccDatabaseDefaultConfig(
					testName,
					redisContainer,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDatabaseExists("qovery_database.test"),
					resource.TestCheckResourceAttr("qovery_database.test", "name", generateTestName(testName)),
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

func testAccDatabaseDefaultConfig(testName string, db database) string {
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), db.dbType, db.version, db.mode,
	)
}

func testAccDatabaseDefaultConfigWithAccessibility(testName string, db database, accessibility string) string {
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  accessibility = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), db.dbType, db.version, db.mode, accessibility,
	)
}

func testAccDatabaseDefaultConfigWithResources(testName string, db database, cpu int64, memory int64) string {
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  cpu = %d
  memory = %d
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), db.dbType, db.version, db.mode, cpu, memory,
	)
}

func testAccDatabaseDefaultConfigWithStorage(testName string, db database, storage int64) string {
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  storage = %d
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), db.dbType, db.version, db.mode, storage,
	)
}

func testAccDatabaseDefaultConfigWithState(testName string, db database, state string) string {
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "%s"
  version = "%s"
  mode = "%s"
  state = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), db.dbType, db.version, db.mode, state,
	)
}
