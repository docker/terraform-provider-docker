package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-dockerhub/internal/pkg/hubclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgSettingRegistryAccessManagement(t *testing.T) {
	orgName := "dockerterraform"
	customRegistry := hubclient.RegistryAccessManagementCustomRegistry{
		Allowed:      true,
		FriendlyName: "My personal registry",
		Address:      "https://example.com",
	}
	altRegistry := hubclient.RegistryAccessManagementCustomRegistry{
		Allowed:      false,
		FriendlyName: "My alt registry",
		Address:      "https://alternate.com",
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, true, true, customRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, true, true, customRegistry),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "dockerhub_org_setting_registry_access_management.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// disable hub
				Config: testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, true, false, customRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// disable ReAM
				Config: testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, false, false, customRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// add custom registry
				Config: testAccOrgSettingRegistryAccessManagementMultipleCustomRegistries(orgName, false, false, customRegistry, altRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "2"),
				),
			},
			{
				// remove custom registries
				Config: testAccOrgSettingRegistryAccessManagementNoCustomRegistry(orgName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("dockerhub_org_setting_registry_access_management.test", "custom_registries.#", "0"),
				),
			},
			{
				// delete
				Config: testAccOrgSettingRegistryAccessManagementBase,
			},
		},
	})
}

const testAccOrgSettingRegistryAccessManagementBase = `
provider "dockerhub" {
  username = "username-placeholder"
  password = "PW"
  host = "https://hub-stage.docker.com/v2"
}`

func testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName string, enabled, allowDockerHub bool, custom hubclient.RegistryAccessManagementCustomRegistry) string {
	return fmt.Sprintf(`
%[1]s

resource "dockerhub_org_setting_registry_access_management" "test" {
  org_name                     = "%[2]s"
  enabled                      = %[3]t
  standard_registry_docker_hub = {
    allowed = %[4]t
  }
  custom_registries = [
    {
	  address       = "%[5]s",
	  friendly_name = "%[6]s",
	  allowed       = %[7]t
  	}
  ]
}
`, testAccOrgSettingRegistryAccessManagementBase,
		orgName,
		enabled,
		allowDockerHub,
		custom.Address,
		custom.FriendlyName,
		custom.Allowed,
	)
}

func testAccOrgSettingRegistryAccessManagementMultipleCustomRegistries(orgName string, enabled, allowDockerHub bool, custom1 hubclient.RegistryAccessManagementCustomRegistry, custom2 hubclient.RegistryAccessManagementCustomRegistry) string {
	return fmt.Sprintf(`
%[1]s

resource "dockerhub_org_setting_registry_access_management" "test" {
  org_name                     = "%[2]s"
  enabled                      = %[3]t
  standard_registry_docker_hub = {
    allowed = %[4]t
  }
  custom_registries = [
    {
	  address       = "%[5]s",
	  friendly_name = "%[6]s",
	  allowed       = %[7]t
  	},
	{
	  address       = "%[8]s",
	  friendly_name = "%[9]s",
	  allowed       = %[10]t
  	}
  ]
}
`,
		testAccOrgSettingRegistryAccessManagementBase,
		orgName,
		enabled,
		allowDockerHub,
		custom1.Address,
		custom1.FriendlyName,
		custom1.Allowed,
		custom2.Address,
		custom2.FriendlyName,
		custom2.Allowed,
	)
}

func testAccOrgSettingRegistryAccessManagementNoCustomRegistry(orgName string, enabled, allowDockerHub bool) string {
	return fmt.Sprintf(`
%[1]s

resource "dockerhub_org_setting_registry_access_management" "test" {
  org_name                     = "%[2]s"
  enabled                      = %[3]t
  standard_registry_docker_hub = {
    allowed = %[4]t
  }
  custom_registries = []
}
`, testAccOrgSettingRegistryAccessManagementBase, orgName, enabled, allowDockerHub)
}
