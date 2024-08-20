package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgTeamMemberDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testOrgTeamMemberDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockerhub_org_team_member.test", "org_name", os.Getenv("DOCKERHUB_ORG_NAME")),
					resource.TestCheckResourceAttr("data.dockerhub_org_team_member.test", "team_name", "example-team"),
					resource.TestCheckResourceAttr("data.dockerhub_org_team_member.test", "members.#", "2"), // assuming there are two members
					resource.TestCheckResourceAttr("data.dockerhub_org_team_member.test", "members.0.username", "user1"),
					resource.TestCheckResourceAttr("data.dockerhub_org_team_member.test", "members.1.username", "user2"),
				),
			},
		},
	})
}

func testOrgTeamMemberDataSourceConfig() string {
	return `
provider "dockerhub" {
  host = "https://hub.docker.com/v2"
}

data "dockerhub_org_team_member" "test" {
  org_name  = "` + os.Getenv("DOCKERHUB_ORG_NAME") + `"
  team_name = "example-team"
}
`
}
