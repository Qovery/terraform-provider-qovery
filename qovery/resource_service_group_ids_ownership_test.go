//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

const (
	testAccServiceLabelsGroupAddress      = "qovery_labels_group.test"
	testAccServiceAnnotationsGroupAddress = "qovery_annotations_group.test"

	// testAccDeclaredServiceGroups makes testAccServiceGroupsConfig attach both groups.
	testAccDeclaredServiceGroups = "__declared__"
)

// testAccServiceGroups is the set of groups attached to a service, by group id.
type testAccServiceGroups struct {
	labels      []string
	annotations []string
}

// testAccServiceGroupsTarget is the service a group ids ownership test runs against.
type testAccServiceGroupsTarget struct {
	address           string
	dataSourceAddress string
	// config renders the service next to a labels group and an annotations group. groupIDs is the
	// HCL value of both labels_group_ids and annotations_group_ids; "" leaves them out.
	config func(testName, groupIDs string, withDataSource bool) string
	// readGroups returns the groups attached to the service in the API.
	readGroups func(serviceID string) (testAccServiceGroups, error)
	// writeGroups attaches exactly these groups through the API, bypassing Terraform, the way the
	// Qovery Console does.
	writeGroups func(serviceID string, groups testAccServiceGroups) error
	exists      resource.TestCheckFunc
	destroy     resource.TestCheckFunc
}

// TestAcc_ApplicationGroupIdsOwnership covers labels_group_ids and annotations_group_ids as desired
// state on the application (the read path shared with qovery_database).
func TestAcc_ApplicationGroupIdsOwnership(t *testing.T) {
	t.Parallel()
	const address = "qovery_application.test"
	testAccServiceGroupIdsOwnership(t, "application-group-ids-ownership", testAccServiceGroupsTarget{
		address:           address,
		dataSourceAddress: "data.qovery_application.test",
		config: func(testName, groupIDs string, withDataSource bool) string {
			return testAccServiceGroupsConfig(testName, groupIDs, withDataSource, "qovery_application", fmt.Sprintf(`
  environment_id  = qovery_environment.test.id
  name            = "%s"
  build_mode      = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url          = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}`, generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID()))
		},
		readGroups:  testAccReadApplicationGroups,
		writeGroups: testAccWriteApplicationGroups,
		exists:      testAccQoveryApplicationExists(address),
		destroy:     testAccQoveryApplicationDestroy(address),
	})
}

// TestAcc_ContainerGroupIdsOwnership covers labels_group_ids and annotations_group_ids as desired
// state on the container (the read path shared with qovery_job).
func TestAcc_ContainerGroupIdsOwnership(t *testing.T) {
	t.Parallel()
	const address = "qovery_container.test"
	testAccServiceGroupIdsOwnership(t, "container-group-ids-ownership", testAccServiceGroupsTarget{
		address:           address,
		dataSourceAddress: "data.qovery_container.test",
		config: func(testName, groupIDs string, withDataSource bool) string {
			return testAccContainerRegistryDefaultConfig(testName) + testAccServiceGroupsConfig(testName, groupIDs, withDataSource, "qovery_container", fmt.Sprintf(`
  environment_id = qovery_environment.test.id
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}`, generateTestName(testName), containerImageName, containerTag))
		},
		readGroups:  testAccReadContainerGroups,
		writeGroups: testAccWriteContainerGroups,
		exists:      testAccQoveryContainerExists(address),
		destroy:     testAccQoveryContainerDestroy(address),
	})
}

// testAccServiceGroupIdsOwnership runs the ownership steps: groups attached or detached outside
// Terraform show in the plan and the next apply reverts them, removing the attributes plans the
// detach, import records the remote groups and the data source reports them. Every step checks
// the groups the API holds, not only the state.
func testAccServiceGroupIdsOwnership(t *testing.T, testName string, target testAccServiceGroupsTarget) {
	const declared = testAccDeclaredServiceGroups
	config := func(groupIDs string, withDataSource bool) string {
		return target.config(testName, groupIDs, withDataSource)
	}
	var serviceID string
	none := testAccServiceGroups{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             target.destroy,
		Steps: []resource.TestStep{
			// 1. Both attributes omitted: stored as null, nothing attached.
			{
				Config: config("", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					target.exists,
					testAccCaptureResourceID(target.address, &serviceID),
					resource.TestCheckNoResourceAttr(target.address, "labels_group_ids.#"),
					resource.TestCheckNoResourceAttr(target.address, "annotations_group_ids.#"),
					testAccCheckServiceGroupsInAPI(target, &serviceID, false),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Groups attached outside Terraform show in the plan while the attributes are omitted.
			{
				Config:             config("", false),
				Check:              testAccSetServiceGroupsOutOfBand(target, &serviceID, true),
				ConfigPlanChecks:   testAccExpectServiceGroupsPlanned(target.address, knownvalue.Null()),
				ExpectNonEmptyPlan: true,
			},
			// 3. The corrective apply detaches them.
			{
				Config:           config("", false),
				Check:            testAccCheckServiceGroupsInAPI(target, &serviceID, false),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Declare both groups.
			{
				Config: config(declared, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(target.address, "labels_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(target.address, "labels_group_ids.*", testAccServiceLabelsGroupAddress, "id"),
					resource.TestCheckResourceAttr(target.address, "annotations_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(target.address, "annotations_group_ids.*", testAccServiceAnnotationsGroupAddress, "id"),
					testAccCheckServiceGroupsInAPI(target, &serviceID, true),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 5. Groups detached outside Terraform show in the plan.
			{
				Config:             config(declared, false),
				Check:              testAccSetServiceGroupsOutOfBand(target, &serviceID, false),
				ConfigPlanChecks:   testAccExpectServiceGroupsPlanned(target.address, knownvalue.SetSizeExact(1)),
				ExpectNonEmptyPlan: true,
			},
			// 6. The corrective apply attaches them again.
			{
				Config:           config(declared, false),
				Check:            testAccCheckServiceGroupsInAPI(target, &serviceID, true),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 7. Import records the attached groups.
			{
				ResourceName:      target.address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 8. The data source reports the attached groups.
			{
				Config: config(declared, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(target.dataSourceAddress, "labels_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(target.dataSourceAddress, "labels_group_ids.*", testAccServiceLabelsGroupAddress, "id"),
					resource.TestCheckResourceAttr(target.dataSourceAddress, "annotations_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(target.dataSourceAddress, "annotations_group_ids.*", testAccServiceAnnotationsGroupAddress, "id"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 9. Removing the attributes plans the detach and the apply detaches.
			{
				Config: config("", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(target.address, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(target.address, tfjsonpath.New("labels_group_ids"), knownvalue.Null()),
						plancheck.ExpectKnownValue(target.address, tfjsonpath.New("annotations_group_ids"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(target.address, "labels_group_ids.#"),
					resource.TestCheckNoResourceAttr(target.address, "annotations_group_ids.#"),
					testAccCheckServiceGroupsInAPI(target, &serviceID, false),
				),
			},
			// 10. An explicit [] keeps the service without groups and plans clean; the data source
			// reports no group as [].
			{
				Config: config("[]", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(target.address, "labels_group_ids.#", "0"),
					resource.TestCheckResourceAttr(target.address, "annotations_group_ids.#", "0"),
					resource.TestCheckResourceAttr(target.dataSourceAddress, "labels_group_ids.#", "0"),
					resource.TestCheckResourceAttr(target.dataSourceAddress, "annotations_group_ids.#", "0"),
					func(_ *terraform.State) error {
						groups, err := target.readGroups(serviceID)
						if err != nil {
							return err
						}
						return testAccCompareServiceGroups(serviceID, groups, none)
					},
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// --- configuration -------------------------------------------------------------------------------

// testAccServiceGroupsConfig renders a service of resourceType with body, next to a labels group,
// an annotations group and the environment they live in. groupIDs testAccDeclaredServiceGroups
// attaches both groups; any other non-empty value is the HCL value of both attributes.
func testAccServiceGroupsConfig(testName, groupIDs string, withDataSource bool, resourceType, body string) string {
	attributes := ""
	switch groupIDs {
	case "":
	case testAccDeclaredServiceGroups:
		attributes = fmt.Sprintf("\n  labels_group_ids      = [%s.id]\n  annotations_group_ids = [%s.id]", testAccServiceLabelsGroupAddress, testAccServiceAnnotationsGroupAddress)
	default:
		attributes = fmt.Sprintf("\n  labels_group_ids      = %s\n  annotations_group_ids = %s", groupIDs, groupIDs)
	}

	dataSource := ""
	if withDataSource {
		dataSource = fmt.Sprintf(`
data "%s" "test" {
  id = %s.test.id
}
`, resourceType, resourceType)
	}

	return fmt.Sprintf(`
%s

resource "qovery_labels_group" "test" {
  organization_id = "%s"
  name            = "%s-lg"
  labels = [{ key = "team", value = "platform", propagate_to_cloud_provider = false }]
}

resource "qovery_annotations_group" "test" {
  organization_id = "%s"
  name            = "%s-ag"
  annotations     = { team = "platform" }
  scopes          = ["PODS"]
}

resource "%s" "test" {%s%s
}
%s`, testAccEnvironmentDefaultConfig(testName),
		getTestOrganizationID(), generateTestName(testName),
		getTestOrganizationID(), generateTestName(testName),
		resourceType, body, attributes, dataSource)
}

// --- plan checks ---------------------------------------------------------------------------------

// testAccExpectServiceGroupsPlanned asserts that the refreshed plan updates the service back to the
// configured group ids after an out-of-band change.
func testAccExpectServiceGroupsPlanned(address string, planned knownvalue.Check) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
			plancheck.ExpectKnownValue(address, tfjsonpath.New("labels_group_ids"), planned),
			plancheck.ExpectKnownValue(address, tfjsonpath.New("annotations_group_ids"), planned),
		},
	}
}

// --- API checks and out-of-band changes ----------------------------------------------------------

// testAccCheckServiceGroupsInAPI checks whether the configuration's labels group and annotations
// group are the only groups attached to the service in the API (attached) or none is.
func testAccCheckServiceGroupsInAPI(target testAccServiceGroupsTarget, serviceID *string, attached bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		want, err := testAccConfigServiceGroups(s, attached)
		if err != nil {
			return err
		}
		got, err := target.readGroups(*serviceID)
		if err != nil {
			return err
		}
		return testAccCompareServiceGroups(*serviceID, got, want)
	}
}

// testAccSetServiceGroupsOutOfBand attaches the configuration's groups (attach) or detaches every
// group through the API, bypassing Terraform.
func testAccSetServiceGroupsOutOfBand(target testAccServiceGroupsTarget, serviceID *string, attach bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groups, err := testAccConfigServiceGroups(s, attach)
		if err != nil {
			return err
		}
		return target.writeGroups(*serviceID, groups)
	}
}

func testAccConfigServiceGroups(s *terraform.State, attached bool) (testAccServiceGroups, error) {
	if !attached {
		return testAccServiceGroups{}, nil
	}
	ids := make([]string, 0, 2)
	for _, address := range []string{testAccServiceLabelsGroupAddress, testAccServiceAnnotationsGroupAddress} {
		rs, ok := s.RootModule().Resources[address]
		if !ok || rs.Primary.ID == "" {
			return testAccServiceGroups{}, fmt.Errorf("%s: id not found in state", address)
		}
		ids = append(ids, rs.Primary.ID)
	}
	return testAccServiceGroups{labels: ids[:1], annotations: ids[1:]}, nil
}

func testAccCompareServiceGroups(serviceID string, got, want testAccServiceGroups) error {
	if fmt.Sprint(got.labels) != fmt.Sprint(want.labels) || fmt.Sprint(got.annotations) != fmt.Sprint(want.annotations) {
		return fmt.Errorf("service %s groups in API: got labels %v annotations %v, want labels %v annotations %v",
			serviceID, got.labels, got.annotations, want.labels, want.annotations)
	}
	return nil
}

func labelsGroupRequests(ids []string) []qovery.ServiceLabelRequest {
	requests := make([]qovery.ServiceLabelRequest, 0, len(ids))
	for _, id := range ids {
		requests = append(requests, qovery.ServiceLabelRequest{Id: id})
	}
	return requests
}

func annotationsGroupRequests(ids []string) []qovery.ServiceAnnotationRequest {
	requests := make([]qovery.ServiceAnnotationRequest, 0, len(ids))
	for _, id := range ids {
		requests = append(requests, qovery.ServiceAnnotationRequest{Id: id})
	}
	return requests
}

func labelsGroupResponseIDs(groups []qovery.OrganizationLabelsGroupResponse) []string {
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.Id)
	}
	return ids
}

func annotationsGroupResponseIDs(groups []qovery.OrganizationAnnotationsGroupResponse) []string {
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.Id)
	}
	return ids
}

func testAccGetApplicationFromAPI(applicationID string) (*qovery.Application, error) {
	app, res, err := qoveryAPIClient.ApplicationMainCallsAPI.GetApplication(context.TODO(), applicationID).Execute()
	if err != nil || (res != nil && res.StatusCode >= 400) {
		return nil, fmt.Errorf("failed to read application %s: %v", applicationID, err)
	}
	return app, nil
}

func testAccReadApplicationGroups(applicationID string) (testAccServiceGroups, error) {
	app, err := testAccGetApplicationFromAPI(applicationID)
	if err != nil {
		return testAccServiceGroups{}, err
	}
	return testAccServiceGroups{labels: labelsGroupResponseIDs(app.LabelsGroups), annotations: annotationsGroupResponseIDs(app.AnnotationsGroups)}, nil
}

// testAccWriteApplicationGroups edits the application with its current settings and the given
// groups. The test application has no storage and no autoscaling policy, so neither is sent.
func testAccWriteApplicationGroups(applicationID string, groups testAccServiceGroups) error {
	app, err := testAccGetApplicationFromAPI(applicationID)
	if err != nil {
		return err
	}
	request := qovery.ApplicationEditRequest{
		Name:                   &app.Name,
		Description:            app.Description,
		BuildMode:              app.BuildMode,
		DockerfilePath:         app.DockerfilePath,
		Cpu:                    app.Cpu,
		Memory:                 app.Memory,
		Gpu:                    app.Gpu,
		EphemeralStorageInGib:  app.EphemeralStorageInGib,
		MinRunningInstances:    app.MinRunningInstances,
		MaxRunningInstances:    app.MaxRunningInstances,
		Healthchecks:           app.Healthchecks,
		AutoPreview:            app.AutoPreview,
		Ports:                  app.Ports,
		Arguments:              app.Arguments,
		Entrypoint:             app.Entrypoint,
		AutoDeploy:             *qovery.NewNullableBool(app.AutoDeploy),
		IconUri:                &app.IconUri,
		DockerTargetBuildStage: app.DockerTargetBuildStage,
		CpuArchitecture:        app.CpuArchitecture,
		LabelsGroups:           labelsGroupRequests(groups.labels),
		AnnotationsGroups:      annotationsGroupRequests(groups.annotations),
	}
	if repo := app.GitRepository; repo != nil {
		request.GitRepository = &qovery.ApplicationGitRepositoryRequest{
			Url:        repo.Url,
			Branch:     repo.Branch,
			RootPath:   repo.RootPath,
			GitTokenId: repo.GitTokenId,
			Provider:   repo.Provider,
		}
	}
	_, res, err := qoveryAPIClient.ApplicationMainCallsAPI.EditApplication(context.TODO(), applicationID).ApplicationEditRequest(request).Execute()
	if err != nil || (res != nil && res.StatusCode >= 400) {
		return fmt.Errorf("failed to edit application %s out of band: %v", applicationID, err)
	}
	return nil
}

func testAccGetContainerFromAPI(containerID string) (*qovery.ContainerResponse, error) {
	cont, res, err := qoveryAPIClient.ContainerMainCallsAPI.GetContainer(context.TODO(), containerID).Execute()
	if err != nil || (res != nil && res.StatusCode >= 400) {
		return nil, fmt.Errorf("failed to read container %s: %v", containerID, err)
	}
	return cont, nil
}

func testAccReadContainerGroups(containerID string) (testAccServiceGroups, error) {
	cont, err := testAccGetContainerFromAPI(containerID)
	if err != nil {
		return testAccServiceGroups{}, err
	}
	return testAccServiceGroups{labels: labelsGroupResponseIDs(cont.LabelsGroups), annotations: annotationsGroupResponseIDs(cont.AnnotationsGroups)}, nil
}

// testAccWriteContainerGroups edits the container with its current settings and the given groups.
// The test container has no ports, no storage and no autoscaling policy, so none is sent.
func testAccWriteContainerGroups(containerID string, groups testAccServiceGroups) error {
	cont, err := testAccGetContainerFromAPI(containerID)
	if err != nil {
		return err
	}
	request := qovery.ContainerRequest{
		Name:                  cont.Name,
		Description:           cont.Description,
		RegistryId:            cont.Registry.Id,
		ImageName:             cont.ImageName,
		Tag:                   cont.Tag,
		Arguments:             cont.Arguments,
		Entrypoint:            cont.Entrypoint,
		Cpu:                   &cont.Cpu,
		Memory:                &cont.Memory,
		Gpu:                   &cont.Gpu,
		EphemeralStorageInGib: cont.EphemeralStorageInGib,
		CpuArchitecture:       cont.CpuArchitecture,
		MinRunningInstances:   &cont.MinRunningInstances,
		MaxRunningInstances:   &cont.MaxRunningInstances,
		Healthchecks:          cont.Healthchecks,
		AutoPreview:           &cont.AutoPreview,
		AutoDeploy:            *qovery.NewNullableBool(cont.AutoDeploy),
		IconUri:               &cont.IconUri,
		LabelsGroups:          labelsGroupRequests(groups.labels),
		AnnotationsGroups:     annotationsGroupRequests(groups.annotations),
	}
	_, res, err := qoveryAPIClient.ContainerMainCallsAPI.EditContainer(context.TODO(), containerID).ContainerRequest(request).Execute()
	if err != nil || (res != nil && res.StatusCode >= 400) {
		return fmt.Errorf("failed to edit container %s out of band: %v", containerID, err)
	}
	return nil
}
