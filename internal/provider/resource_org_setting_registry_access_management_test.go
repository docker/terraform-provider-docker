package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/docker/terraform-provider-docker/internal/hubclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgSettingRegistryAccessManagement(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
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
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, true, true, customRegistry),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_registry_access_management.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// disable hub
				Config: testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, true, false, customRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// disable ReAM
				Config: testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName, false, false, customRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "1"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.address", "https://example.com"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.friendly_name", "My personal registry"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.0.allowed", "true"),
				),
			},
			{
				// add custom registry
				Config: testAccOrgSettingRegistryAccessManagementMultipleCustomRegistries(orgName, false, false, customRegistry, altRegistry),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "2"),
				),
			},
			{
				// remove custom registries
				Config: testAccOrgSettingRegistryAccessManagementNoCustomRegistry(orgName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "standard_registry_docker_hub.allowed", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_registry_access_management.test", "custom_registries.#", "0"),
				),
			},
			{
				// delete
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingRegistryAccessManagementCustomRegistry(orgName string, enabled, allowDocker bool, custom hubclient.RegistryAccessManagementCustomRegistry) string {
	return fmt.Sprintf(`
resource "docker_org_setting_registry_access_management" "test" {
  org_name                     = "%[1]s"
  enabled                      = %[2]t
  standard_registry_docker_hub = {
    allowed = %[3]t
  }
  custom_registries = [
    {
	  address       = "%[4]s",
	  friendly_name = "%[5]s",
	  allowed       = %[6]t
  	}
  ]
}
`, orgName, enabled, allowDocker, custom.Address, custom.FriendlyName, custom.Allowed)
}

func testAccOrgSettingRegistryAccessManagementMultipleCustomRegistries(orgName string, enabled, allowDocker bool, custom1 hubclient.RegistryAccessManagementCustomRegistry, custom2 hubclient.RegistryAccessManagementCustomRegistry) string {
	return fmt.Sprintf(`
resource "docker_org_setting_registry_access_management" "test" {
  org_name                     = "%[1]s"
  enabled                      = %[2]t
  standard_registry_docker_hub = {
    allowed = %[3]t
  }
  custom_registries = [
    {
	  address       = "%[4]s",
	  friendly_name = "%[5]s",
	  allowed       = %[6]t
  	},
	{
	  address       = "%[7]s",
	  friendly_name = "%[8]s",
	  allowed       = %[9]t
  	}
  ]
}
`,
		orgName,
		enabled,
		allowDocker,
		custom1.Address,
		custom1.FriendlyName,
		custom1.Allowed,
		custom2.Address,
		custom2.FriendlyName,
		custom2.Allowed,
	)
}

func testAccOrgSettingRegistryAccessManagementNoCustomRegistry(orgName string, enabled, allowDocker bool) string {
	return fmt.Sprintf(`
resource "docker_org_setting_registry_access_management" "test" {
  org_name                     = "%[1]s"
  enabled                      = %[2]t
  standard_registry_docker_hub = {
    allowed = %[3]t
  }
  custom_registries = []
}
`, orgName, enabled, allowDocker)
}
