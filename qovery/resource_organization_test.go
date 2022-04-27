package qovery_test

// TODO: to remove as we don't handle create / delete for organization
//func TestAcc_Organization(t *testing.T) {
//	t.Parallel()
//	organizationNameSuffix := uuid.New().String()
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryOrganizationDestroy("qovery_organization.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccOrganizationConfig(
//					generateOrganizationName(organizationNameSuffix),
//					"FREE",
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryOrganizationExists("qovery_organization.test"),
//					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateOrganizationName(organizationNameSuffix)),
//					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
//					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
//				),
//			},
//			// Update name
//			{
//				Config: testAccOrganizationConfig(
//					generateOrganizationName(fmt.Sprintf("updated-%s", organizationNameSuffix)),
//					"FREE",
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateOrganizationName(fmt.Sprintf("updated-%s", organizationNameSuffix))),
//					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
//					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
//				),
//			},
//			// Add description
//			{
//				Config: testAccOrganizationConfigWithDescription(
//					generateOrganizationName(fmt.Sprintf("updated-%s", organizationNameSuffix)),
//					"FREE",
//					"this is my description",
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateOrganizationName(fmt.Sprintf("updated-%s", organizationNameSuffix))),
//					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
//					resource.TestCheckResourceAttr("qovery_organization.test", "description", "this is my description"),
//				),
//			},
//		},
//	})
//}

// TODO: avoid creating an organization to test the import since we can't create an organization
//func TestAcc_OrganizationImport(t *testing.T) {
//	t.Parallel()
//	organizationNameSuffix := uuid.New().String()
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryOrganizationDestroy("qovery_organization.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccOrganizationConfig(
//					generateOrganizationName(organizationNameSuffix),
//					"FREE",
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryOrganizationExists("qovery_organization.test"),
//					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateOrganizationName(organizationNameSuffix)),
//					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
//					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
//				),
//			},
//			// Check Import
//			{
//				ResourceName:      "qovery_organization.test",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}

//func testAccQoveryOrganizationExists(resourceName string) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[resourceName]
//		if !ok {
//			return fmt.Errorf("organization not found: %s", resourceName)
//		}
//
//		if rs.Primary.ID == "" {
//			return fmt.Errorf("organization.id not found")
//		}
//
//		_, err := apiClient.GetOrganization(context.TODO(), rs.Primary.ID)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//}
//
//func testAccQoveryOrganizationDestroy(resourceName string) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[resourceName]
//		if !ok {
//			return fmt.Errorf("organization not found: %s", resourceName)
//		}
//
//		if rs.Primary.ID == "" {
//			return fmt.Errorf("organization.id not found")
//		}
//
//		_, apiErr := apiClient.GetOrganization(context.TODO(), rs.Primary.ID)
//		if apiErr == nil {
//			// TODO: handle orga delete properly
//			// return fmt.Errorf("found organization but expected it to have been deleted")
//			return nil
//		}
//		if !apierrors.IsNotFound(apiErr) {
//			return fmt.Errorf("unexpected error checking for deleted organization: %s", apiErr.Summary())
//		}
//		return nil
//	}
//}
//
//func generateOrganizationName(suffix string) string {
//	return fmt.Sprintf("%s-organization-%s", testNamePrefix, suffix)
//}
//
//func testAccOrganizationConfig(name string, plan string) string {
//	return fmt.Sprintf(`
//resource "qovery_organization" "test" {
//  name = "%s"
//  plan = "%s"
//}
//`, name, plan)
//}
//
//func testAccOrganizationConfigWithDescription(name string, plan string, description string) string {
//	return fmt.Sprintf(`
//resource "qovery_organization" "test" {
//  name = "%s"
//  plan = "%s"
//  description = "%s"
//}
//`, name, plan, description)
//}
