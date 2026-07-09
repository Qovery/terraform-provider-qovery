//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Out-of-band removal tests for domain-service resources (QOV-2030 Track B).
//
// Same "disappears" shape as resource_read_removed_out_of_band_test.go, but for resources
// whose Read goes through a domain service returning a plain error: the resource is deleted
// out-of-band via the raw generated API client, then the post-apply refresh must drop it
// from state (handleDomainReadNotFound → RemoveResource) instead of erroring, and
// ExpectNonEmptyPlan captures the resulting re-create plan.
//
// project and environment are the ticket-prioritized representatives that are cheap to
// provision (pure API objects, no cloud deploy). The remaining Track B resources share the
// exact same helper and wrap chain (service errors.Wrap around a domain apierrors.APIError),
// covered by the unit tests in read_not_found_domain_test.go.

func TestAcc_ProjectRemovedOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "project-out-of-band"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
				),
			},
			{
				Config: testAccProjectDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDisappearsViaRawAPI("qovery_project.test",
						func(id string) error {
							_, err := qoveryAPIClient.ProjectMainCallsAPI.DeleteProject(context.TODO(), id).Execute()
							return err
						},
						func(id string) int {
							_, res, _ := qoveryAPIClient.ProjectMainCallsAPI.GetProject(context.TODO(), id).Execute()
							return rawStatusCode(res)
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc_EnvironmentRemovedOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "environment-out-of-band"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
				),
			},
			{
				Config: testAccEnvironmentDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDisappearsViaRawAPI("qovery_environment.test",
						func(id string) error {
							_, err := qoveryAPIClient.EnvironmentMainCallsAPI.DeleteEnvironment(context.TODO(), id).Execute()
							return err
						},
						func(id string) int {
							_, res, _ := qoveryAPIClient.EnvironmentMainCallsAPI.GetEnvironment(context.TODO(), id).Execute()
							return rawStatusCode(res)
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Deliberately NOT t.Parallel(): see TestAcc_CustomRole — role churn races q-core's
// project_role_permission matrix maintenance (unlocked cross-entity inserts), causing flaky
// FK-violation 500s in every concurrently-running project-creating test.
func TestAcc_CustomRoleRemovedOutOfBand(t *testing.T) {
	orgID := getTestOrganizationID()
	roleName := generateTestName("custom-role-out-of-band")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoleConfigNamed(roleName, "DEPLOYER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
				),
			},
			{
				Config: testAccCustomRoleConfigNamed(roleName, "DEPLOYER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDisappearsViaRawAPI("qovery_custom_role.test",
						func(id string) error {
							_, err := qoveryAPIClient.OrganizationCustomRoleAPI.DeleteOrganizationCustomRole(context.TODO(), orgID, id).Execute()
							return err
						},
						func(id string) int {
							_, res, _ := qoveryAPIClient.OrganizationCustomRoleAPI.GetOrganizationCustomRole(context.TODO(), orgID, id).Execute()
							return rawStatusCode(res)
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// rawStatusCode extracts an HTTP status code from a raw generated-client response.
// The generated client returns an error alongside the response for non-2xx statuses,
// so callers intentionally ignore that error and rely on the status code alone.
// A nil response (network error) is reported as 0 (unknown, keep polling).
func rawStatusCode(res *http.Response) int {
	if res == nil {
		return 0
	}
	return res.StatusCode
}

// testAccDisappearsViaRawAPI deletes the resource out-of-band using the raw generated API
// client, then polls the raw GET until the API reports it gone (404/403), so the post-apply
// refresh deterministically sees the deleted state even when the API deletes asynchronously
// (e.g. environments go through a deletion pipeline).
func testAccDisappearsViaRawAPI(resourceName string, del func(id string) error, getStatus func(id string) int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("%s: id not found in state", resourceName)
		}
		if err := del(rs.Primary.ID); err != nil {
			return fmt.Errorf("%s: failed to delete out-of-band: %s", resourceName, err)
		}
		// Wait until the API reports the resource gone (bounded).
		for attempt := 0; attempt < 60; attempt++ {
			status := getStatus(rs.Primary.ID)
			if status == http.StatusNotFound || status == http.StatusForbidden {
				return nil
			}
			time.Sleep(2 * time.Second)
		}
		return fmt.Errorf("%s: still present after out-of-band delete (timeout)", resourceName)
	}
}
