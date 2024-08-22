package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgMemberAssociationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testOrgMemberAssociationResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_member_association.test", "invite_id"),
					resource.TestCheckResourceAttr("docker_org_member_association.test", "org_name", "dockerhackathon"),
					resource.TestCheckResourceAttr("docker_org_member_association.test", "team_name", "test"),
					resource.TestCheckResourceAttr("docker_org_member_association.test", "user_name", "newtest@example.com"),
					resource.TestCheckResourceAttr("docker_org_member_association.test", "role", "member"),
				),
			},
			// {
			// 	ResourceName:            "docker_org_member_association.test",
			// 	ImportState:             false,
			// 	ImportStateVerify:       false,
			// 	ImportStateVerifyIgnore: []string{"invite_id"},
			// },
		},
	})
}

func testOrgMemberAssociationResourceConfig() string {
	return `
provider "docker" {
}

resource "docker_org_member_association" "test" {
  org_name  = "dockerhackathon"
  team_name = "test"
  user_name = "newtest@example.com"
  role = "member"
}
`
}
