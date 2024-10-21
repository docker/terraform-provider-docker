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

func TestAccOrgTeamDataSource(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	teamName := "teamname"
	description := "Test org team description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgTeamDataSourceConfig(orgName, teamName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.docker_org_team.test", "org_name", orgName),
					resource.TestCheckResourceAttr("data.docker_org_team.test", "team_name", teamName),
					resource.TestCheckResourceAttrSet("data.docker_org_team.test", "id"),
					resource.TestCheckResourceAttrSet("data.docker_org_team.test", "uuid"),
					resource.TestCheckResourceAttr("data.docker_org_team.test", "description", description),
					resource.TestCheckResourceAttr("data.docker_org_team.test", "member_count", "0"),
				),
			},
		},
	})
}

func testAccOrgTeamDataSourceConfig(orgName string, teamName string, description string) string {
	return fmt.Sprintf(`

resource "docker_org_team" "terraform-team" {
  org_name         = "%s"
  team_name        = "%s"
  team_description = "%s"
}

data "docker_org_team" "test" {
  org_name  = docker_org_team.terraform-team.org_name
  team_name = docker_org_team.terraform-team.team_name
}
`, orgName, teamName, description)
}
