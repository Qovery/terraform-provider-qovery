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
	"github.com/qovery/terraform-provider-qovery/qovery"
)

// Out-of-band removal tests for client-layer resources.
//
// These "disappears" acceptance tests delete the resource out-of-band (directly via the
// Qovery API, bypassing Terraform) and then assert that the next refresh/plan does NOT
// error: the resource must be dropped from state (handleReadNotFound → RemoveResource) so
// Terraform plans a re-create. `ExpectNonEmptyPlan: true` captures exactly that — the
// post-apply refresh removes the resource and the plan becomes non-empty (a re-create).
//
// Before the fix, the out-of-band delete made Read return a hard diagnostic, so the
// post-apply refresh errored and the step failed.
//
// Coverage note: cluster_dns_provider is the fourth client-layer resource. It has no standalone
// acceptance scaffolding and no independent delete (it lives inside a cluster), so it is not
// given a bespoke disappears test here. It calls the same handleReadNotFound helper as the
// resources below, and that helper is exhaustively covered by the unit test in
// read_not_found_test.go.

// TestAcc_DatabaseRemovedOutOfBand is runnable in the normal loop: a REDIS container
// database is cheap and fast to provision.
func TestAcc_DatabaseRemovedOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "database-out-of-band"
	dbModel := qovery.Database{
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
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckRemovedOutOfBand("qovery_database.test", func(id string) *apierrors.APIError {
			_, apiErr := apiClient.GetDatabase(context.TODO(), id)
			return apiErr
		}),
		Steps: []resource.TestStep{
			{
				Config: GetDatabaseConfigFromModel(testName, dbModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDatabaseExists("qovery_database.test"),
				),
			},
			{
				// Delete the database out-of-band, then expect the next plan to re-create it.
				Config: GetDatabaseConfigFromModel(testName, dbModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDisappears("qovery_database.test", func(id string) *apierrors.APIError {
						return apiClient.DeleteDatabase(context.TODO(), id)
					}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAcc_ApplicationRemovedOutOfBand is runnable in the normal loop.
func TestAcc_ApplicationRemovedOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "application-out-of-band"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckRemovedOutOfBand("qovery_application.test", func(id string) *apierrors.APIError {
			_, apiErr := apiClient.GetApplication(context.TODO(), id, "{}", false)
			return apiErr
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
				),
			},
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDisappears("qovery_application.test", func(id string) *apierrors.APIError {
						return apiClient.DeleteApplication(context.TODO(), id)
					}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAcc_ClusterRemovedOutOfBand provisions a real Kubernetes cluster (slow + costly);
// intended for CI, not the interactive loop. This is the originally reported resource.
func TestAcc_ClusterRemovedOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "cluster-out-of-band"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckRemovedOutOfBand("qovery_cluster.test", func(id string) *apierrors.APIError {
			_, apiErr := apiClient.GetCluster(context.TODO(), getTestOrganizationID(), id, "{}", false)
			return apiErr
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigWithKeda(testName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
				),
			},
			{
				Config: testAccClusterConfigWithKeda(testName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDisappears("qovery_cluster.test", func(id string) *apierrors.APIError {
						return apiClient.DeleteCluster(context.TODO(), getTestOrganizationID(), id)
					}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccQoveryDisappears deletes a resource out-of-band via the given delete closure,
// mirroring the get-closure shape already used by testAccCheckRemovedOutOfBand.
func testAccQoveryDisappears(resourceName string, del func(id string) *apierrors.APIError) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("%s: id not found in state", resourceName)
		}
		if apiErr := del(rs.Primary.ID); apiErr != nil {
			return fmt.Errorf("%s: failed to delete out-of-band: %s", resourceName, apiErr.Detail())
		}
		return nil
	}
}

// testAccCheckRemovedOutOfBand is the CheckDestroy for the "disappears" tests. The resource
// is deleted out-of-band mid-test and then dropped from Terraform state by handleReadNotFound,
// so its absence from state is the expected clean end-state — not a dangling-resource failure.
// When the resource is still tracked in state, it confirms the API reports it as deleted.
func testAccCheckRemovedOutOfBand(resourceName string, get func(id string) *apierrors.APIError) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary.ID == "" {
			return nil
		}
		apiErr := get(rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("%s: found resource but expected it to be deleted", resourceName)
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("%s: unexpected error checking deletion: %s", resourceName, apiErr.Summary())
		}
		return nil
	}
}
