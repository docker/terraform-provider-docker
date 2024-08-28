package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgMemberResource(t *testing.T) {
	org := envvar.GetWithDefault(envvar.AccTestOrganization)
	teamName := "test" + randString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testOrgMemberResourceConfig(org, teamName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_member.test", "invite_id"),
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", org),
					resource.TestCheckResourceAttr("docker_org_member.test", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_member.test", "user_name", "newtest@example.com"),
					resource.TestCheckResourceAttr("docker_org_member.test", "role", "member"),
				),
			},
			// {
			// 	ResourceName:            "docker_org_member.test",
			// 	ImportState:             false,
			// 	ImportStateVerify:       false,
			// 	ImportStateVerifyIgnore: []string{"invite_id"},
			// },
		},
	})
}

func testOrgMemberResourceConfig(org string, team string) string {
	return fmt.Sprintf(`
provider "docker" {
  host = "hub-stage.docker.com"
}

resource "docker_org_team" "testing" {
  org_name         = "%[1]s"
  team_name        = "%[2]s"
}

resource "docker_org_member" "test" {
  org_name  = docker_org_team.testing.org_name
  team_name = docker_org_team.testing.team_name
  user_name = "newtest@example.com"
  role = "member"
}`, org, team)
}
