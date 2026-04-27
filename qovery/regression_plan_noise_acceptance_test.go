//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/qovery/terraform-provider-qovery/qovery"
)

func TestAcc_PlanNoise_Application(t *testing.T) {
	t.Parallel()
	testName := "plan-noise-application"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
				),
			},
			{
				Config:             testAccApplicationDefaultConfig(testName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAcc_PlanNoise_Database(t *testing.T) {
	t.Parallel()
	testName := "plan-noise-database"
	dbConfig := GetDatabaseConfigFromModel(
		testName,
		qovery.Database{
			Name:          qovery.FromString(generateTestName(testName)),
			IconUri:       qovery.FromString(fmt.Sprintf("app://qovery-console/%s", generateTestName(testName))),
			Type:          qovery.FromString("REDIS"),
			Version:       qovery.FromString("6.2"),
			Mode:          qovery.FromString("CONTAINER"),
			Accessibility: qovery.FromString("PUBLIC"),
			CPU:           qovery.FromInt32(250),
			Memory:        qovery.FromInt32(256),
			Storage:       qovery.FromInt32(10),
			InstanceType:  qovery.FromStringPointer(nil),
		},
	)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDatabaseDestroy("qovery_database.test"),
		Steps: []resource.TestStep{
			{
				Config: dbConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
				),
			},
			{
				Config:             dbConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAcc_PlanNoise_Environment(t *testing.T) {
	t.Parallel()
	testName := "plan-noise-environment"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
				),
			},
			{
				Config:             testAccEnvironmentDefaultConfig(testName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
