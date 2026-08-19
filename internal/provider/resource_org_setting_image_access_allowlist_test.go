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

func TestAccOrgSettingImageAccessAllowlist(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create with two repos
				Config: testAccOrgSettingImageAccessAllowlist(orgName, []string{
					orgName + "/repo-a",
					orgName + "/repo-b",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_allowlist.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_image_access_allowlist.test", "repositories.#", "2"),
					resource.TestCheckTypeSetElemAttr("docker_org_setting_image_access_allowlist.test", "repositories.*", orgName+"/repo-a"),
					resource.TestCheckTypeSetElemAttr("docker_org_setting_image_access_allowlist.test", "repositories.*", orgName+"/repo-b"),
				),
			},
			{
				// import
				Config: testAccOrgSettingImageAccessAllowlist(orgName, []string{
					orgName + "/repo-a",
					orgName + "/repo-b",
				}),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_image_access_allowlist.test",
			},
			{
				// add one, remove one
				Config: testAccOrgSettingImageAccessAllowlist(orgName, []string{
					orgName + "/repo-b",
					orgName + "/repo-c",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_image_access_allowlist.test", "repositories.#", "2"),
					resource.TestCheckTypeSetElemAttr("docker_org_setting_image_access_allowlist.test", "repositories.*", orgName+"/repo-b"),
					resource.TestCheckTypeSetElemAttr("docker_org_setting_image_access_allowlist.test", "repositories.*", orgName+"/repo-c"),
				),
			},
			{
				// delete (removes all items)
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingImageAccessAllowlist(orgName string, repositories []string) string {
	repoList := ""
	for _, r := range repositories {
		repoList += fmt.Sprintf("    %q,\n", r)
	}
	return fmt.Sprintf(`
resource "docker_org_setting_namespace" "test" {
  org_name                       = %[1]q
  disable_public_repositories    = false
  disable_push_member_namespaces = false
  repository_allowlist_enabled   = true
}

resource "docker_org_setting_image_access_allowlist" "test" {
  org_name     = docker_org_setting_namespace.test.org_name
  repositories = [
%[2]s  ]
}
`, orgName, repoList)
}
