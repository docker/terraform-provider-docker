package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgTeamResource(t *testing.T) {
	orgName := "dockerterraform"
	teamName := "test" + randString(10)
	updatedName := "test" + randString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// create
				Config: testAccOrgTeamResource(orgName, teamName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_team.testing", "id"),
					resource.TestCheckResourceAttr("docker_org_team.testing", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_description", "test description"),
				),
			},
			{
				// import
				Config:        testAccOrgTeamResource(orgName, teamName, "test description"),
				ImportState:   true,
				ImportStateId: orgName + "/" + teamName,
				ResourceName:  "docker_org_team.testing",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_team.testing", "id"),
					resource.TestCheckResourceAttr("docker_org_team.testing", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_description", "test description"),
				),
			},
			{
				// update description
				Config: testAccOrgTeamResource(orgName, teamName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_team.testing", "id"),
					resource.TestCheckResourceAttr("docker_org_team.testing", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_name", teamName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_description", "updated description"),
				),
			},
			{
				// update team name
				Config: testAccOrgTeamResource(orgName, updatedName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_team.testing", "id"),
					resource.TestCheckResourceAttr("docker_org_team.testing", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_name", updatedName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_description", "updated description"),
				),
			},
			{
				// delete
				Config: testAccOrgTeamResourceBase,
			},
			{
				// create no description
				Config: testAccOrgTeamResourceNoDescription(orgName, teamName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("docker_org_team.testing", "id"),
					resource.TestCheckResourceAttr("docker_org_team.testing", "org_name", orgName),
					resource.TestCheckResourceAttr("docker_org_team.testing", "team_name", teamName),
					resource.TestCheckNoResourceAttr("docker_org_team.testing", "team_description"),
				),
			},
		},
	})
}

const testAccOrgTeamResourceBase = `
provider "docker" {
  host = "hub-stage.docker.com"
}
`

func testAccOrgTeamResource(orgName, teamName, teamDesc string) string {
	return fmt.Sprintf(`
%s
resource "docker_org_team" "testing" {
  org_name         = "%s"
  team_name        = "%s"
  team_description = "%s"
}
`, testAccOrgTeamResourceBase, orgName, teamName, teamDesc)
}

func testAccOrgTeamResourceNoDescription(orgName, teamName string) string {
	return fmt.Sprintf(`
%s
resource "docker_org_team" "testing" {
  org_name         = "%s"
  team_name        = "%s"
}
`, testAccOrgTeamResourceBase, orgName, teamName)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
