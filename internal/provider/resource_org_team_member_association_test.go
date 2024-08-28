package provider

import (
	"fmt"
	"testing"

	"github.com/docker/terraform-provider-docker/internal/envvar"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrgTeamMemberAssociation(t *testing.T) {
	orgName := envvar.GetWithDefault(envvar.AccTestOrganization)
	teamName := "testteam"
	userName := "dockerterraformprovideracctest"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgTeamMemberAssociationConfig(orgName, teamName, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("docker_org_team_member_association.test", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team_member_association.test", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_team_member_association.test", "user_name", userName),
				),
			},
			{
				Config: " ",
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceDoesNotExist("docker_org_team_member_association.test"),
				),
			},
		},
	})
}

func testAccOrgTeamMemberAssociationConfig(orgName, teamName, userName string) string {
	return fmt.Sprintf(`
resource "docker_org_team" "test" {
  org_name   = "%[1]s"
  team_name  = "%[2]s"
}

resource "docker_org_team_member_association" "test" {
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
