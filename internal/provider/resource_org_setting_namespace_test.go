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

func TestAccOrgSettingNamespace(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccOrgSettingNamespace(orgName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingNamespace(orgName, true, true),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_namespace.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
				),
			},
			{
				// update
				Config: testAccOrgSettingNamespace(orgName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
				),
			},
			{
				// delete (resets to false/false)
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingNamespace(orgName string, disablePublicRepositories, disablePushMemberNamespaces bool) string {
	return fmt.Sprintf(`
resource "docker_org_setting_namespace" "test" {
  org_name                       = "%[1]s"
  disable_public_repositories    = %[2]t
  disable_push_member_namespaces = %[3]t
}
`, orgName, disablePublicRepositories, disablePushMemberNamespaces)
}
