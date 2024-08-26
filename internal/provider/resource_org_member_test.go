package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgMemberResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testOrgMemberResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_member.test", "invite_id"),
					resource.TestCheckResourceAttr("docker_org_member.test", "org_name", "dockerhackathon"),
					resource.TestCheckResourceAttr("docker_org_member.test", "team_name", "test"),
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

func testOrgMemberResourceConfig() string {
	return `
provider "docker" {
  host = "hub-stage.docker.com"
}

resource "docker_org_member" "test" {
  org_name  = "dockerhackathon"
  team_name = "test"
  user_name = "newtest@example.com"
  role = "member"
}
`
}
