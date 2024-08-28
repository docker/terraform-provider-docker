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
