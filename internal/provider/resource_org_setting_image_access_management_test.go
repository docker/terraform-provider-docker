package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgSettingImageAccessManagement(t *testing.T) {
	orgName := "dockerterraform"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccOrgSettingImageAccessManagement(orgName, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_verified_publishers", "false"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingImageAccessManagement(orgName, true, false, false),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "dockerhub_org_setting_image_access_management.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_verified_publishers", "false"),
				),
			},
			{
				// update setting
				Config: testAccOrgSettingImageAccessManagement(orgName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_verified_publishers", "true"),
				),
			},
			{
				// disable iam
				Config: testAccOrgSettingImageAccessManagement(orgName, false, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_official_images", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_image_access_management.test", "allow_verified_publishers", "true"),
				),
			},
			{
				// delete
				Config: testAccOrgSettingImageAccessManagementBase,
			},
		},
	})
}

const testAccOrgSettingImageAccessManagementBase = `
provider "dockerhub" {
  host = "https://hub-stage.docker.com/v2"
}`

func testAccOrgSettingImageAccessManagement(orgName string, enabled, allowOfficialImages, allowVerifiedPublishers bool) string {
	return fmt.Sprintf(`
%[1]s

resource "dockerhub_org_setting_image_access_management" "test" {
  org_name                  = "%[2]s"
  enabled                   = %[3]t
  allow_official_images     = %[4]t
  allow_verified_publishers = %[5]t
}
`, testAccOrgSettingImageAccessManagementBase, orgName, enabled, allowOfficialImages, allowVerifiedPublishers)
}
