/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgSettingRegistry(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create with private visibility
				Config: testAccOrgSettingRegistry(orgName, "private"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "default_repo_visibility", "private"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingRegistry(orgName, "private"),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_registry.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "default_repo_visibility", "private"),
				),
			},
			{
				// update to public
				Config: testAccOrgSettingRegistry(orgName, "public"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_registry.test", "default_repo_visibility", "public"),
				),
			},
			{
				// delete (resets to public)
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingRegistry(orgName, defaultRepoVisibility string) string {
	return fmt.Sprintf(`
resource "docker_org_setting_registry" "test" {
  org_name                = %q
  default_repo_visibility = %q
}
`, orgName, defaultRepoVisibility)
}
