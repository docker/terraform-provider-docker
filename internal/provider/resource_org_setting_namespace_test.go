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
				Config: testAccOrgSettingNamespace(orgName, true, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "repository_allowlist_enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "require_federated_auth_for_push", "true"),
				),
			},
			{
				// import
				Config:        testAccOrgSettingNamespace(orgName, true, true, true, true),
				ImportState:   true,
				ImportStateId: orgName,
				ResourceName:  "docker_org_setting_namespace.test",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "repository_allowlist_enabled", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "require_federated_auth_for_push", "true"),
				),
			},
			{
				// update
				Config: testAccOrgSettingNamespace(orgName, false, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "true"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "repository_allowlist_enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "require_federated_auth_for_push", "false"),
				),
			},
			{
				// omitting the optional settings falls back to the false defaults
				Config: testAccOrgSettingNamespaceDefaults(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_public_repositories", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "disable_push_member_namespaces", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "repository_allowlist_enabled", "false"),
					resource.TestCheckResourceAttr("docker_org_setting_namespace.test", "require_federated_auth_for_push", "false"),
				),
			},
			{
				// delete (resets all to false)
				Config: " ",
			},
		},
	})
}

func testAccOrgSettingNamespace(orgName string, disablePublicRepositories, disablePushMemberNamespaces, repositoryAllowlistEnabled, requireFederatedAuthForPush bool) string {
	return fmt.Sprintf(`
resource "docker_org_setting_namespace" "test" {
  org_name                        = "%[1]s"
  disable_public_repositories     = %[2]t
  disable_push_member_namespaces  = %[3]t
  repository_allowlist_enabled    = %[4]t
  require_federated_auth_for_push = %[5]t
}
`, orgName, disablePublicRepositories, disablePushMemberNamespaces, repositoryAllowlistEnabled, requireFederatedAuthForPush)
}

func testAccOrgSettingNamespaceDefaults(orgName string) string {
	return fmt.Sprintf(`
resource "docker_org_setting_namespace" "test" {
  org_name = "%[1]s"
}
`, orgName)
}
