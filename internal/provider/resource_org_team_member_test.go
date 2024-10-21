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
	"os"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrgTeamMember(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	teamName := fmt.Sprintf("test%s", randString(5))
	userName := os.Getenv("DOCKER_USERNAME")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgTeamMemberConfig(orgName, teamName, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_team_member.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team_member.test", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_team_member.test", "user_name", userName),
				),
			},
			{
				Config: " ",
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceDoesNotExist("docker_org_team_member.test"),
				),
			},
		},
	})
}

func testAccOrgTeamMemberConfig(orgName, teamName, userName string) string {
	return fmt.Sprintf(`
resource "docker_org_team" "test" {
  org_name   = "%[1]s"
  team_name  = "%[2]s"
}

resource "docker_org_team_member" "test" {
  org_name   = docker_org_team.test.org_name
  team_name  = docker_org_team.test.team_name
  user_name  = "%[3]s"
}
`, orgName, teamName, userName)
}

func testCheckResourceDoesNotExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[resourceName]; ok {
			return fmt.Errorf("Resource %s still exists", resourceName)
		}
		return nil
	}
}
