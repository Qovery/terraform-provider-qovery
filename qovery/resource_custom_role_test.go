//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// Deliberately NOT t.Parallel(): custom role create/update/delete races q-core's
// project_role_permission matrix maintenance. Project creation inserts permission rows for
// every existing custom role (and role creation for every existing project) without locking,
// so running this alongside any project-creating test yields flaky FK-violation 500s.
// Serial tests run while all parallel tests are paused, which removes the overlap entirely.
func TestAcc_CustomRole(t *testing.T) {
	roleName := generateTestName("custom-role")
	var roleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryCustomRoleDestroy("qovery_custom_role.test"),
		Steps: []resource.TestStep{
			// Step 1: reserved name rejected at plan time (placed first so no state dangles)
			{
				Config:      testAccCustomRoleConfigNamed("admin", "MANAGER"),
				ExpectError: regexp.MustCompile(`reserved`),
			},
			// Step 1b: is_admin=true with an explicit (even empty) permissions set is rejected
			// at plan time (placed before any state-creating step so nothing dangles).
			{
				Config:      testAccCustomRoleConfigAdminWithEmptyPermissions(roleName),
				ExpectError: regexp.MustCompile(`is_admin`),
			},
			// Step 2: create with a declared project permission (4 env types)
			{
				Config: testAccCustomRoleConfigNamed(roleName, "DEPLOYER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "name", roleName),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.#", "1"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.0.permissions.#", "4"),
					testAccCaptureResourceID("qovery_custom_role.test", &roleID),
				),
			},
			// Step 3: update a permission in place (PRODUCTION DEPLOYER -> MANAGER)
			{
				Config: testAccCustomRoleConfigNamed(roleName, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.#", "1"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.0.permissions.#", "4"),
				),
			},
			// Step 4: dropping the `description` attribute from config plans its reset to the ""
			// q-core stores for an omitted description, and the apply clears it.
			{
				Config: testAccCustomRoleConfigNoDescription(roleName, "MANAGER"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("qovery_custom_role.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("qovery_custom_role.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "description", ""),
				),
			},
			// Step 5: adding an unrelated project must NOT produce a diff on the role
			// (THE perpetual-diff regression test: the server matrix now includes the new
			// project with default perms, which the Read must filter out).
			{
				Config:             testAccCustomRoleConfigWithExtraProject(roleName, "MANAGER"),
				Check:              testAccQoveryCustomRoleExists("qovery_custom_role.test"),
				ExpectNonEmptyPlan: false,
			},
			// Step 6: a permission granted from the Console on an undeclared cluster shows up in
			// the plan...
			{
				Config: testAccCustomRoleConfigWithExtraProject(roleName, "MANAGER"),
				Check: func(_ *terraform.State) error {
					return testAccSetCustomRoleClusterPermissionOutOfBand(roleID, getTestClusterID(), "ADMIN")
				},
				ExpectNonEmptyPlan: true,
			},
			// ... the refresh stores it...
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_custom_role.test", "cluster_permissions.#", "1"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "cluster_permissions.0.cluster_id", getTestClusterID()),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "cluster_permissions.0.permission", "ADMIN"),
				),
				ExpectNonEmptyPlan: true,
			},
			// ... and the next apply resets it to the VIEWER default.
			{
				Config: testAccCustomRoleConfigWithExtraProject(roleName, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qovery_custom_role.test", "cluster_permissions.#"),
					testAccCheckCustomRoleClusterPermission(&roleID, getTestClusterID(), "VIEWER"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Step 7: import keeps non-default entries (id format: "org_id,role_id")
			{
				ResourceName:      "qovery_custom_role.test",
				ImportState:       true,
				ImportStateIdFunc: testAccCustomRoleImportStateID("qovery_custom_role.test"),
				ImportStateVerify: true,
			},
		},
	})
}

// testAccSetCustomRoleClusterPermissionOutOfBand sets the permission of the role on a cluster,
// bypassing Terraform. The edit resends the whole matrix, as q-core requires.
func testAccSetCustomRoleClusterPermissionOutOfBand(roleID, clusterID, permission string) error {
	apiPath := fmt.Sprintf("/organization/%s/customRole/%s", getTestOrganizationID(), roleID)
	found := false
	err := testAccEditServiceOutOfBand(apiPath, nil, func(role map[string]any) {
		clusters, _ := role["cluster_permissions"].([]any)
		for _, c := range clusters {
			entry, ok := c.(map[string]any)
			if ok && entry["cluster_id"] == clusterID {
				entry["permission"] = permission
				found = true
			}
		}
	})
	if err == nil && !found {
		return fmt.Errorf("cluster %s is not in the matrix of custom role %s", clusterID, roleID)
	}
	return err
}

// testAccCheckCustomRoleClusterPermission checks the permission of the role on a cluster in the API.
// roleID is read when the check runs, since an earlier step captures it.
func testAccCheckCustomRoleClusterPermission(roleID *string, clusterID, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		role, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), *roleID)
		if err != nil {
			return err
		}
		for _, cp := range role.ClusterPermissions {
			if cp.ClusterID == clusterID {
				if string(cp.Permission) != expected {
					return fmt.Errorf("custom role %s has %s on cluster %s, expected %s", *roleID, cp.Permission, clusterID, expected)
				}
				return nil
			}
		}
		return fmt.Errorf("cluster %s is not in the matrix of custom role %s", clusterID, *roleID)
	}
}

func testAccCustomRoleConfigNamed(roleName string, prodPermission string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"
  description     = "acceptance test role"

  project_permissions = [
    {
      project_id = "%s"
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "%s" },
      ]
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID(), prodPermission)
}

func testAccCustomRoleConfigAdminWithEmptyPermissions(roleName string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"
  description     = "acceptance test role"

  project_permissions = [
    {
      project_id  = "%s"
      is_admin    = true
      permissions = []
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID())
}

func testAccCustomRoleConfigNoDescription(roleName string, prodPermission string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"

  project_permissions = [
    {
      project_id = "%s"
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "%s" },
      ]
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID(), prodPermission)
}

// testAccCustomRoleConfigWithExtraProject adds a project to the configuration of step 4, so the
// role does not change in the apply that creates the project: q-core rejects a role update that
// races a project creation (the matrix it reads misses the new project).
func testAccCustomRoleConfigWithExtraProject(roleName string, prodPermission string) string {
	return testAccCustomRoleConfigNoDescription(roleName, prodPermission) + fmt.Sprintf(`
resource "qovery_project" "extra" {
  organization_id = "%s"
  name            = "%s-extra"
}
`, getTestOrganizationID(), roleName)
}

func testAccCustomRoleImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("custom role not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.ID), nil
	}
}

func testAccQoveryCustomRoleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("custom role not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("custom_role.id not found")
		}

		_, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryCustomRoleDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("custom role not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("custom_role.id not found")
		}

		_, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found custom role but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted custom role: %s", err.Error())
		}
		return nil
	}
}

// Serial like TestAcc_CustomRole (see the comment there).
func TestAcc_CustomRoleRemoteAccess(t *testing.T) {
	roleName := generateTestName("custom-role-ra")
	var roleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryCustomRoleDestroy("qovery_custom_role.test"),
		Steps: []resource.TestStep{
			// shell or port_forward cannot be true with NO_ACCESS, rejected at plan time
			{
				Config:      testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "NO_ACCESS", shell = true }`, "DEPLOYER"),
				ExpectError: regexp.MustCompile(`NO_ACCESS`),
			},
			// declared flags are sent, the others follow the level
			{
				Config: testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "NO_ACCESS" }`, "DEPLOYER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					testAccCaptureResourceID("qovery_custom_role.test", &roleID),
					testAccCheckCustomRoleRemoteAccess(&roleID, "DEVELOPMENT", false, true),
					testAccCheckCustomRoleRemoteAccess(&roleID, "PREVIEW", true, true),
					testAccCheckCustomRoleRemoteAccess(&roleID, "STAGING", true, false),
					testAccCheckCustomRoleRemoteAccess(&roleID, "PRODUCTION", false, false),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// a shell revoked from the console on an environment type where the flag is not set
			// is not drift: Terraform doesn't manage it
			{
				Config: testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "NO_ACCESS" }`, "DEPLOYER"),
				Check: func(_ *terraform.State) error {
					return testAccSetCustomRoleShellOutOfBand(roleID, "PREVIEW", false)
				},
			},
			// ... and it survives a change of the level, as the revocation was deliberate
			{
				Config: testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "NO_ACCESS" }`, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoleRemoteAccess(&roleID, "PREVIEW", false, true),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// raising an environment type that was never customised follows the level
			{
				Config: testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "DEPLOYER" }`, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoleRemoteAccess(&roleID, "PRODUCTION", true, true),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// the data source reports the flags
			{
				Config: testAccCustomRoleConfigRemoteAccess(roleName, `{ environment_type = "PRODUCTION", permission = "DEPLOYER" }`, "MANAGER") + `
data "qovery_custom_role" "test" {
  organization_id = qovery_custom_role.test.organization_id
  id              = qovery_custom_role.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoleDataSourceRemoteAccess("data.qovery_custom_role.test", "DEVELOPMENT", "false", "true"),
					testAccCheckCustomRoleDataSourceRemoteAccess("data.qovery_custom_role.test", "PREVIEW", "false", "true"),
					testAccCheckCustomRoleDataSourceRemoteAccess("data.qovery_custom_role.test", "STAGING", "true", "false"),
					testAccCheckCustomRoleDataSourceRemoteAccess("data.qovery_custom_role.test", "PRODUCTION", "true", "true"),
				),
			},
		},
	})
}

// testAccCheckCustomRoleDataSourceRemoteAccess checks shell and port_forward the data source reports for an
// environment type of the test project. The nested sets are keyed by hash, so the entry is found by its values.
func testAccCheckCustomRoleDataSourceRemoteAccess(resourceName, envType, shell, portForward string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}
		attrs := rs.Primary.Attributes
		for key, value := range attrs {
			if !strings.HasSuffix(key, ".environment_type") || value != envType {
				continue
			}
			entry := strings.TrimSuffix(key, ".environment_type")
			projectKey := entry[:strings.Index(entry, ".permissions.")]
			if attrs[projectKey+".project_id"] != getTestProjectID() {
				continue
			}
			if attrs[entry+".shell"] != shell || attrs[entry+".port_forward"] != portForward {
				return fmt.Errorf("%s: shell=%s port_forward=%s, expected shell=%s port_forward=%s",
					envType, attrs[entry+".shell"], attrs[entry+".port_forward"], shell, portForward)
			}
			return nil
		}
		return fmt.Errorf("%s of the test project is not in %s", envType, resourceName)
	}
}

// testAccCustomRoleConfigRemoteAccess has DEVELOPMENT with shell off and port_forward on, PREVIEW at
// previewPermission with the flags not set, STAGING with shell on, and the given PRODUCTION entry.
func testAccCustomRoleConfigRemoteAccess(roleName, productionEntry, previewPermission string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"

  project_permissions = [
    {
      project_id = "%s"
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER", shell = false, port_forward = true },
        { environment_type = "PREVIEW", permission = "%s" },
        { environment_type = "STAGING", permission = "VIEWER", shell = true },
        %s,
      ]
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID(), previewPermission, productionEntry)
}

// testAccCheckCustomRoleRemoteAccess checks shell and port_forward of the role on the test project.
func testAccCheckCustomRoleRemoteAccess(roleID *string, envType string, shell, portForward bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		role, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), *roleID)
		if err != nil {
			return err
		}
		for _, pp := range role.ProjectPermissions {
			if pp.ProjectID != getTestProjectID() {
				continue
			}
			for _, ep := range pp.Permissions {
				if string(ep.EnvironmentType) != envType {
					continue
				}
				gotShell, gotPortForward := ep.Shell != nil && *ep.Shell, ep.PortForward != nil && *ep.PortForward
				if gotShell != shell || gotPortForward != portForward {
					return fmt.Errorf("custom role %s on %s: shell=%t port_forward=%t, expected shell=%t port_forward=%t",
						*roleID, envType, gotShell, gotPortForward, shell, portForward)
				}
				return nil
			}
		}
		return fmt.Errorf("%s is not in the matrix of custom role %s", envType, *roleID)
	}
}

// testAccSetCustomRoleShellOutOfBand sets shell of the role on an environment type of the test
// project, bypassing Terraform. The edit resends the whole matrix, as q-core requires.
func testAccSetCustomRoleShellOutOfBand(roleID, envType string, shell bool) error {
	apiPath := fmt.Sprintf("/organization/%s/customRole/%s", getTestOrganizationID(), roleID)
	found := false
	err := testAccEditServiceOutOfBand(apiPath, nil, func(role map[string]any) {
		projects, _ := role["project_permissions"].([]any)
		for _, p := range projects {
			project, ok := p.(map[string]any)
			if !ok || project["project_id"] != getTestProjectID() {
				continue
			}
			permissions, _ := project["permissions"].([]any)
			for _, e := range permissions {
				entry, ok := e.(map[string]any)
				if ok && entry["environment_type"] == envType {
					entry["shell"] = shell
					found = true
				}
			}
		}
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("custom role %s: %s is not in the matrix", roleID, envType)
	}
	return nil
}
